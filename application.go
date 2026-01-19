package web

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sort"
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
	paramsPool    sync.Pool
	maxParams     uint16
	globalAllowed []string

	NotFound http.Handler
}

// New return *web.Application
func New() *Application {
	app := &Application{}
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

// Get method
func (app *Application) Get(path string, next Next) {
	app.addRoute(http.MethodGet, path, next)
}

// Head method
func (app *Application) Head(path string, cb Next) {
	app.addRoute(http.MethodHead, path, cb)
}

// Post method
func (app *Application) Post(path string, next Next) {
	app.addRoute(http.MethodPost, path, next)
}

// Put method
func (app *Application) Put(path string, next Next) {
	app.addRoute(http.MethodPut, path, next)
}

// Patch method
func (app *Application) Patch(path string, next Next) {
	app.addRoute(http.MethodPatch, path, next)
}

// Delete method
func (app *Application) Delete(path string, next Next) {
	app.addRoute(http.MethodDelete, path, next)
}

// Options method
func (app *Application) Options(path string, next Next) {
	app.addRoute(http.MethodOptions, path, next)
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

	if root := app.trees[r.Method]; root != nil {

		if next, params, _ := root.getValue(rel, app.getParams); next != nil {

			c := createCtx(w, r, params)
			defer releaseCtx(c)

			val, err := next(c)

			app.putParams(params)

			if err != nil {

				code := errCode(err)

				if e, ok := err.(*errFn); ok {
					if err := e.cb(w, r); err != nil {
						writeHeader(w, r, code)
						c.write(err.Error())
						app.Errf("%s %s %d %s %s %d %v", r.RemoteAddr, r.Host, c.UserId(), r.Method, rel, code, err)
					}
					return
				}

				writeHeader(w, r, code)
				c.write(err.Error())

				app.Errf("%s %s %d %s %s %d %v", r.RemoteAddr, r.Host, c.UserId(), r.Method, rel, code, err)

				return
			}

			if val != nil {

				code := http.StatusOK

				switch r.Method {
				case http.MethodPost:
					code = http.StatusCreated
				}

				writeHeader(w, r, code)
				c.write(val)

				app.Logf("%s %s %d %s %s %d", r.RemoteAddr, r.Host, c.UserId(), r.Method, rel, code)

				if rel, ok := val.(IRelease); ok {
					rel.Release()
				}
			} else {
				writeHeader(w, r, http.StatusNoContent)

				app.Logf("%s %s %d %s %s %d", r.RemoteAddr, r.Host, c.UserId(), r.Method, rel, 204)
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
		writeHeader(w, r, http.StatusNoContent)
		return
	}

	if app.NotFound != nil {
		app.NotFound.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
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

	if app.paramsPool.New == nil && app.maxParams > 0 {
		app.paramsPool.New = func() any {
			ps := make(Params, 0, app.maxParams)
			return &ps
		}
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
	*ps = (*ps)[0:0] // reset slice
	return ps
}

func (app *Application) putParams(ps *Params) {
	if ps != nil {
		app.paramsPool.Put(ps)
	}
}

func (app *Application) recv(w http.ResponseWriter, r *http.Request) {
	if rcv := recover(); rcv != nil {
		writeHeader(w, r, http.StatusInternalServerError)
		if app.panic != nil {
			app.panic(w, r, rcv)
		} else {
			app.Errf("%s %s %s %s rcv: %v", r.RemoteAddr, r.Host, r.Method, r.URL.Path, rcv)
		}
	}
}
