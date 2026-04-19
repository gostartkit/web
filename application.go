package web

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// Application is type of a web.Application
type Application struct {
	srv           *http.Server
	trees         map[string]*node
	info          *log.Logger
	err           *log.Logger
	cors          Cors
	panic         Panic
	errorHandler  ErrorHandler
	middleware    Chain
	readers       [mediaTypeSlots]Reader
	writers       [mediaTypeSlots]Writer
	hasReaders    bool
	hasWriters    bool
	paramsPool    sync.Pool
	maxParams     uint16
	globalAllowed []string

	NotFound         http.Handler
	MethodNotAllowed http.Handler
}

// New return *web.Application
func New() *Application {
	app := &Application{}
	app.paramsPool.New = func() any {
		n := app.maxParams
		if n == 0 {
			n = 1
		}
		ps := make(Params, 0, n)
		return &ps
	}
	return app
}

// SetInfoLogger set info logger
func (app *Application) SetInfoLogger(logger *log.Logger) {
	app.info = logger
}

// SetErrLogger set err logger
func (app *Application) SetErrLogger(logger *log.Logger) {
	app.err = logger
}

// SetCORS set CORS
func (app *Application) SetCORS(cors Cors) {
	app.cors = cors
}

// SetPanic set Panic
func (app *Application) SetPanic(panic Panic) {
	app.panic = panic
}

// SetErrorHandler sets a custom route error handler.
func (app *Application) SetErrorHandler(handler ErrorHandler) {
	app.errorHandler = handler
}

// RegisterReader registers a request body reader for a supported content type.
func (app *Application) RegisterReader(contentType string, reader Reader) error {
	mt := parseMediaType(contentType)
	if mt == mediaUnknown {
		return ErrContentType
	}
	app.readers[mt] = reader
	app.hasReaders = true
	return nil
}

// RegisterWriter registers a response writer for a supported accept/content type.
func (app *Application) RegisterWriter(contentType string, writer Writer) error {
	mt := parseMediaType(contentType)
	if mt == mediaUnknown {
		return ErrContentType
	}
	app.writers[mt] = writer
	app.hasWriters = true
	return nil
}

// Use appends application middleware for subsequently registered routes.
func (app *Application) Use(middleware ...Middleware) {
	app.middleware = append(app.middleware, middleware...)
}

// Group creates a route group with a shared prefix and middleware chain.
func (app *Application) Group(prefix string, middleware ...Middleware) *RouteGroup {
	if prefix != "" && prefix[0] != '/' {
		panic("group prefix must begin with '/' in path '" + prefix + "'")
	}
	return &RouteGroup{
		app:        app,
		prefix:     prefix,
		middleware: append(Chain(nil), middleware...),
	}
}

// Handle registers a route for an arbitrary HTTP method.
func (app *Application) Handle(method string, path string, next Next, middleware ...Middleware) {
	app.addRoute(method, path, wrapNext(next, app.middleware, Chain(middleware)))
}

// Get method
func (app *Application) Get(path string, next Next) {
	app.Handle(http.MethodGet, path, next)
}

// Head method
func (app *Application) Head(path string, cb Next) {
	app.Handle(http.MethodHead, path, cb)
}

// Post method
func (app *Application) Post(path string, next Next) {
	app.Handle(http.MethodPost, path, next)
}

// Put method
func (app *Application) Put(path string, next Next) {
	app.Handle(http.MethodPut, path, next)
}

// Patch method
func (app *Application) Patch(path string, next Next) {
	app.Handle(http.MethodPatch, path, next)
}

// Delete method
func (app *Application) Delete(path string, next Next) {
	app.Handle(http.MethodDelete, path, next)
}

// Options method
func (app *Application) Options(path string, next Next) {
	app.Handle(http.MethodOptions, path, next)
}

func (app *Application) addRoute(method string, path string, next Next) {

	if method == "" {
		panic("method must not be empty")
	}

	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if next == nil {
		panic("callback must not be nil")
	}

	if app.trees == nil {
		app.trees = make(map[string]*node)
	}

	root := app.trees[method]

	if root == nil {
		root = new(node)
		app.trees[method] = root
		app.globalAllowed = app.allowed("*", "")
	}

	root.addRoute(path, next)

	if pc := countParams(path); pc > app.maxParams {
		app.maxParams = pc
	}
}

// ServeFiles registers a route to serve static files from the specified file system under the given path pattern.
// The path must end with "/*filepath" to capture file paths dynamically.
// It panics if the path pattern is invalid, ensuring correct configuration during initialization.
//
// Parameters:
//   - path: The URL path pattern (e.g., "/static/*filepath") to match file requests.
//     Must end with "/*filepath" to extract the file path as a parameter.
//   - root: The http.FileSystem to serve files from (e.g., http.Dir("./static")).
//
// Panics:
//   - If the path is shorter than 10 characters or does not end with "/*filepath".
func (app *Application) ServeFiles(path string, root http.FileSystem) {
	// Validate the path pattern to ensure it ends with "/*filepath" for dynamic file path capturing.
	// This check prevents incorrect routing configurations.
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	app.Get(path, func(c *Ctx) (any, error) {
		c.r.URL.Path = c.Param("filepath")
		fileServer.ServeHTTP(c.w, c.r)
		return nil, nil
	})
}

// ServeHTTP w, r
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer app.recv(w, r)

	rel := r.URL.Path
	infoLogger := app.info
	errLogger := app.err

	if root := app.trees[r.Method]; root != nil {

		if next, params, _ := root.getValue(rel, app.getParams); next != nil {

			c := createCtx(app, w, r, params)
			val, err := next(c)
			userID := c.UserId()

			if err != nil {
				code, writeErr := app.handleError(c, err)
				app.putParams(params)
				releaseCtx(c)
				if writeErr != nil && errLogger != nil {
					errLogger.Printf("%s %s %d %s %s %d write error: %v", r.RemoteAddr, r.Host, userID, r.Method, rel, code, writeErr)
				}
				if errLogger != nil {
					errLogger.Printf("%s %s %d %s %s %d %v", r.RemoteAddr, r.Host, userID, r.Method, rel, code, err)
				}

				return
			}

			if val != nil {
				code := statusFromResult(c, val, nil)
				if !c.responseCommitted {
					writeCodeByMedia(w, c.responseMediaType(), code)
				}
				err := c.write(val)
				app.putParams(params)
				releaseCtx(c)
				if err != nil {
					if errLogger != nil {
						errLogger.Printf("%s %s %d %s %s %d write error: %v", r.RemoteAddr, r.Host, userID, r.Method, rel, code, err)
					}
					return
				}

				if infoLogger != nil {
					infoLogger.Printf("%s %s %d %s %s %d", r.RemoteAddr, r.Host, userID, r.Method, rel, code)
				}

				if rel, ok := val.(IRelease); ok {
					rel.Release()
				}
			} else {
				code := statusFromResult(c, nil, nil)
				committed := c.responseCommitted
				app.putParams(params)
				releaseCtx(c)
				if !committed {
					w.WriteHeader(code)
				}

				if infoLogger != nil {
					infoLogger.Printf("%s %s %d %s %s %d", r.RemoteAddr, r.Host, userID, r.Method, rel, code)
				}
			}

			return
		}
	}

	if r.Method == http.MethodOptions && app.cors != nil {
		// Handle OPTIONS requests
		if allow := app.allowed(rel, http.MethodOptions); len(allow) > 0 {
			if origin := r.Header.Get("Origin"); origin != "" {
				app.cors(w.Header().Set, origin, allow)
			}
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if allow := app.allowed(rel, r.Method); len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		if app.MethodNotAllowed != nil {
			app.MethodNotAllowed.ServeHTTP(w, r)
		} else {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
		return
	}

	if app.NotFound != nil {
		app.NotFound.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (app *Application) handleError(c *Ctx, err error) (int, error) {
	if app.errorHandler != nil {
		if nextErr := app.errorHandler(c, err); nextErr == nil {
			return errCode(err), nil
		} else {
			err = nextErr
		}
	}

	code := errCode(err)
	if e, ok := err.(*errFn); ok {
		if cbErr := e.cb(c.w, c.r); cbErr == nil {
			return code, nil
		} else {
			err = cbErr
			code = errCode(err)
		}
	}

	writeCodeByMedia(c.w, c.responseMediaType(), code)
	return code, c.write(err.Error())
}

func wrapNext(next Next, chains ...Chain) Next {
	for i := len(chains) - 1; i >= 0; i-- {
		chain := chains[i]
		for j := len(chain) - 1; j >= 0; j-- {
			if mw := chain[j]; mw != nil {
				next = mw(next)
			}
		}
	}
	return next
}

func joinPaths(prefix, path string) string {
	switch {
	case prefix == "":
		return path
	case path == "":
		return prefix
	case prefix[len(prefix)-1] == '/' && path[0] == '/':
		return prefix[:len(prefix)-1] + path
	case prefix[len(prefix)-1] != '/' && path[0] != '/':
		return prefix + "/" + path
	default:
		return prefix + path
	}
}

// Use appends middleware to the group.
func (g *RouteGroup) Use(middleware ...Middleware) {
	g.middleware = append(g.middleware, middleware...)
}

// Group creates a nested route group.
func (g *RouteGroup) Group(prefix string, middleware ...Middleware) *RouteGroup {
	if prefix != "" && prefix[0] != '/' {
		panic("group prefix must begin with '/' in path '" + prefix + "'")
	}
	child := &RouteGroup{
		app:        g.app,
		prefix:     joinPaths(g.prefix, prefix),
		middleware: append(append(Chain(nil), g.middleware...), middleware...),
	}
	return child
}

// Handle registers a route on the group.
func (g *RouteGroup) Handle(method string, path string, next Next, middleware ...Middleware) {
	g.app.addRoute(method, joinPaths(g.prefix, path), wrapNext(next, g.app.middleware, g.middleware, Chain(middleware)))
}

// Get registers a GET route on the group.
func (g *RouteGroup) Get(path string, next Next) {
	g.Handle(http.MethodGet, path, next)
}

// Head registers a HEAD route on the group.
func (g *RouteGroup) Head(path string, next Next) {
	g.Handle(http.MethodHead, path, next)
}

// Post registers a POST route on the group.
func (g *RouteGroup) Post(path string, next Next) {
	g.Handle(http.MethodPost, path, next)
}

// Put registers a PUT route on the group.
func (g *RouteGroup) Put(path string, next Next) {
	g.Handle(http.MethodPut, path, next)
}

// Patch registers a PATCH route on the group.
func (g *RouteGroup) Patch(path string, next Next) {
	g.Handle(http.MethodPatch, path, next)
}

// Delete registers a DELETE route on the group.
func (g *RouteGroup) Delete(path string, next Next) {
	g.Handle(http.MethodDelete, path, next)
}

// Options registers an OPTIONS route on the group.
func (g *RouteGroup) Options(path string, next Next) {
	g.Handle(http.MethodOptions, path, next)
}

func (app *Application) allowed(path, reqMethod string) []string {

	allowed := make([]string, 0, 9)

	if path == "*" { // server-wide
		// empty method is used for internal calls to refresh the cache
		if reqMethod == "" {
			for method := range app.trees {
				if method == http.MethodOptions {
					continue
				}
				// Add request method to list of allowed methods
				allowed = append(allowed, method)
			}
		} else {
			return app.globalAllowed
		}
	} else { // specific path
		for method := range app.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == http.MethodOptions {
				continue
			}

			cb, _, _ := app.trees[method].getValue(path, nil)
			if cb != nil {
				// Add request method to list of allowed methods
				allowed = append(allowed, method)
			}
		}
	}

	if len(allowed) > 0 {

		allowed = append(allowed, http.MethodOptions)

		sort.Strings(allowed)
	}

	return allowed
}

// ListenAndServe Serve with options on addr
func (app *Application) ListenAndServe(network string, addr string, fns ...func(*http.Server)) error {

	l, err := net.Listen(network, addr)

	if err != nil {
		return err
	}

	defer l.Close()

	return app.serve(l, fns...)
}

// ListenAndServeTLS Serve with tls and options on addr
func (app *Application) ListenAndServeTLS(network string, addr string, tlsConfig *tls.Config, fns ...func(*http.Server)) error {

	l, err := tls.Listen(network, addr, tlsConfig)

	if err != nil {
		return err
	}

	defer l.Close()

	return app.serve(l, fns...)
}

// Shutdown gracefully shuts down the HTTP server, allowing active connections to complete.
// It returns an error if the server is not initialized, ensuring callers can detect invalid states.
//
// Parameters:
//   - ctx: Context controlling the shutdown timeout. If the context expires before shutdown completes,
//     remaining connections may be forcibly closed.
//
// Returns:
//   - error: Returns ErrServerNotInitialized if the server is not initialized,
//     or any error from http.Server.Shutdown (e.g., context timeout).
//     Returns nil if shutdown completes successfully.
func (app *Application) Shutdown(ctx context.Context) error {
	if app.srv == nil {
		return ErrServerNotInitialized
	}
	return app.srv.Shutdown(ctx)
}

func (app *Application) serve(listener net.Listener, fns ...func(*http.Server)) error {

	mux := http.NewServeMux()

	mux.Handle("/", app)

	app.srv = &http.Server{
		Handler: mux,
	}

	for _, fn := range fns {
		fn(app.srv)
	}

	if err := app.srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Inspect method
func (app *Application) Inspect() string {
	return ""
}

// Logf write info log
func (app *Application) Logf(format string, v ...any) {
	if app.info != nil {
		app.info.Printf(format, v...)
	}
}

// Errf write err log
func (app *Application) Errf(format string, v ...any) {
	if app.err != nil {
		app.err.Printf(format, v...)
	}
}

func (app *Application) getParams() *Params {
	ps := app.paramsPool.Get().(*Params)
	if uint16(cap(*ps)) < app.maxParams {
		*ps = make(Params, 0, app.maxParams)
	} else {
		*ps = (*ps)[0:0]
	}
	return ps
}

func (app *Application) putParams(ps *Params) {
	if ps != nil {
		app.paramsPool.Put(ps)
	}
}

func (app *Application) recv(w http.ResponseWriter, r *http.Request) {
	if rcv := recover(); rcv != nil {
		writeCode(w, r, http.StatusInternalServerError)
		if app.panic != nil {
			app.panic(w, r, rcv)
		} else {
			app.Errf("%s %s %s %s rcv: %v", r.RemoteAddr, r.Host, r.Method, r.URL.Path, rcv)
		}
	}
}
