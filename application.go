package web

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const methodRootSlots = 9

func methodRootIndex(method string) int {
	switch method {
	case http.MethodGet:
		return 0
	case http.MethodHead:
		return 1
	case http.MethodPost:
		return 2
	case http.MethodPut:
		return 3
	case http.MethodPatch:
		return 4
	case http.MethodDelete:
		return 5
	case http.MethodOptions:
		return 6
	case http.MethodConnect:
		return 7
	case http.MethodTrace:
		return 8
	default:
		return -1
	}
}

// Application is the HTTP router, middleware registry, and request dispatcher.
//
// Routes are registered before serving and are matched by HTTP method plus path.
// Application implements http.Handler, so it can be passed directly to
// http.Server, httptest, or any standard library integration point.
type Application struct {
	srv             *http.Server
	trees           map[string]*node
	methodRoots     [methodRootSlots]*node
	frozenTrees     map[string]*frozenNode
	frozenRoots     [methodRootSlots]*frozenNode
	info            *log.Logger
	err             *log.Logger
	cors            Cors
	panic           Panic
	errorHandler    ErrorHandler
	middleware      Chain
	readers         [mediaTypeSlots]Reader
	writers         [mediaTypeSlots]Writer
	hasReaders      bool
	hasWriters      bool
	paramsPool      sync.Pool
	paramValuesPool sync.Pool
	maxParams       uint16
	globalAllowed   []string

	NotFound         http.Handler
	MethodNotAllowed http.Handler
}

// New returns a configured *web.Application.
//
// Options are applied immediately, before any routes are registered. Calling New
// without options is equivalent to the historical zero-configuration setup.
//
// Example:
//
//	app := web.New(
//		web.WithMiddleware(web.RequestID("", nil), web.Recover(nil)),
//		web.WithErrorHandler(web.JSONErrorHandler(true)),
//	)
func New(options ...Option) *Application {
	app := &Application{}
	app.paramsPool.New = func() any {
		n := app.maxParams
		if n == 0 {
			n = 1
		}
		ps := make(Params, 0, n)
		return &ps
	}
	app.paramValuesPool.New = func() any {
		n := app.maxParams
		if n == 0 {
			n = 1
		}
		values := make([]string, 0, n)
		return &values
	}
	for _, option := range options {
		if option != nil {
			option(app)
		}
	}
	return app
}

// WithInfoLogger configures the informational logger used by the application.
//
// The logger is used for successful framework-managed route completions. Passing
// nil disables informational logging. Framework-emitted request logs use logfmt
// key/value fields.
func WithInfoLogger(logger *log.Logger) Option {
	return func(app *Application) {
		app.info = logger
	}
}

// WithErrLogger configures the error logger used by the application.
//
// The logger is used for route errors, write errors, and recovered panics handled
// by Application.recv. Passing nil disables error logging. Framework-emitted
// error logs use logfmt key/value fields.
func WithErrLogger(logger *log.Logger) Option {
	return func(app *Application) {
		app.err = logger
	}
}

// WithCORS configures the CORS hook used for automatic OPTIONS responses.
//
// The hook runs only for OPTIONS requests handled by the framework's automatic
// method discovery path.
func WithCORS(cors Cors) Option {
	return func(app *Application) {
		app.cors = cors
	}
}

// WithPanic configures the panic hook used by the recovery boundary in ServeHTTP.
//
// This hook is for panics that reach the Application-level recover boundary. For
// route-level panic handling that participates in normal error writing, prefer
// Recover or RecoverWithOptions middleware.
func WithPanic(panic Panic) Option {
	return func(app *Application) {
		app.panic = panic
	}
}

// WithErrorHandler configures the application-level route error handler.
//
// The handler is called when a route returns a non-nil error. Returning nil from
// the handler means the error response has already been written.
func WithErrorHandler(handler ErrorHandler) Option {
	return func(app *Application) {
		app.errorHandler = handler
	}
}

// WithMiddleware appends application middleware for subsequently registered routes.
//
// Middleware is still applied at route registration time. This option is a
// declarative shortcut for calling app.Use immediately after New.
func WithMiddleware(middleware ...Middleware) Option {
	return func(app *Application) {
		app.Use(middleware...)
	}
}

// WithNotFound configures the handler used when no route matches.
//
// When handler is nil, Application falls back to http.NotFound.
func WithNotFound(handler http.Handler) Option {
	return func(app *Application) {
		app.NotFound = handler
	}
}

// WithMethodNotAllowed configures the handler used when a path exists for other methods.
//
// The framework sets the Allow header before invoking the handler. When handler
// is nil, Application falls back to http.Error with status 405.
func WithMethodNotAllowed(handler http.Handler) Option {
	return func(app *Application) {
		app.MethodNotAllowed = handler
	}
}

// SetInfoLogger sets the informational logger used after successful requests.
//
// This setter is equivalent to constructing the application with WithInfoLogger.
// Framework-emitted request logs use logfmt key/value fields.
func (app *Application) SetInfoLogger(logger *log.Logger) {
	app.info = logger
}

// SetErrLogger sets the logger used for route, write, and panic errors.
//
// This setter is equivalent to constructing the application with WithErrLogger.
// Framework-emitted error logs use logfmt key/value fields.
func (app *Application) SetErrLogger(logger *log.Logger) {
	app.err = logger
}

// SetCORS sets the CORS hook used by automatic OPTIONS responses.
//
// The hook receives a header setter, the request Origin, and the methods allowed
// for the requested route.
func (app *Application) SetCORS(cors Cors) {
	app.cors = cors
}

// SetPanic sets the application-level panic hook.
//
// Prefer Recover or RecoverWithOptions middleware for route-level panic recovery.
func (app *Application) SetPanic(panic Panic) {
	app.panic = panic
}

// SetErrorHandler sets a custom route error handler.
//
// The handler is called for errors returned by route handlers. Return nil when
// the handler has written the response; return an error to let the default writer
// handle it.
func (app *Application) SetErrorHandler(handler ErrorHandler) {
	app.errorHandler = handler
}

// RegisterReader registers a request body reader for a supported content type.
//
// The content type must map to one of the framework's known media types, such as
// application/json or application/xml. A registered reader overrides the built-in
// decoder for TryParseBody.
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
//
// The content type must map to one of the framework's known media types. A
// registered writer overrides the built-in response writer for framework-managed
// responses negotiated to that media type.
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
//
// Middleware is applied when each route is registered. Calling Use does not
// affect routes that were already registered.
func (app *Application) Use(middleware ...Middleware) {
	app.middleware = append(app.middleware, middleware...)
}

// Group creates a route group with a shared prefix and middleware chain.
//
// Routes registered on the returned group inherit the application's current
// middleware plus the group middleware. The prefix must be empty or start with /.
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
//
// The path must start with /. Optional route middleware is applied after
// application and group middleware, and only affects this route.
func (app *Application) Handle(method string, path string, next Next, middleware ...Middleware) {
	app.addRoute(method, path, wrapNext(next, app.middleware, Chain(middleware)))
}

// HTTPHandler adapts a standard net/http handler into a web route handler.
//
// The adapted handler writes directly to the Ctx, which implements
// http.ResponseWriter. It returns (nil, nil) because the response is considered
// manually handled.
func HTTPHandler(handler http.Handler) Next {
	if handler == nil {
		return nil
	}
	return func(c *Ctx) (any, error) {
		handler.ServeHTTP(c, c.r)
		return nil, nil
	}
}

// HandleHTTP registers a standard net/http handler for an arbitrary HTTP method.
//
// This is useful for mounting existing handlers from the standard library or
// third-party packages without rewriting them as Next handlers.
func (app *Application) HandleHTTP(method string, path string, handler http.Handler, middleware ...Middleware) {
	app.Handle(method, path, HTTPHandler(handler), middleware...)
}

// Get registers a route for HTTP GET requests.
func (app *Application) Get(path string, next Next) {
	app.Handle(http.MethodGet, path, next)
}

// GetHTTP registers a standard net/http handler for HTTP GET requests.
func (app *Application) GetHTTP(path string, handler http.Handler) {
	app.HandleHTTP(http.MethodGet, path, handler)
}

// Head registers a route for HTTP HEAD requests.
func (app *Application) Head(path string, cb Next) {
	app.Handle(http.MethodHead, path, cb)
}

// HeadHTTP registers a standard net/http handler for HTTP HEAD requests.
func (app *Application) HeadHTTP(path string, handler http.Handler) {
	app.HandleHTTP(http.MethodHead, path, handler)
}

// Post registers a route for HTTP POST requests.
func (app *Application) Post(path string, next Next) {
	app.Handle(http.MethodPost, path, next)
}

// PostHTTP registers a standard net/http handler for HTTP POST requests.
func (app *Application) PostHTTP(path string, handler http.Handler) {
	app.HandleHTTP(http.MethodPost, path, handler)
}

// Put registers a route for HTTP PUT requests.
func (app *Application) Put(path string, next Next) {
	app.Handle(http.MethodPut, path, next)
}

// PutHTTP registers a standard net/http handler for HTTP PUT requests.
func (app *Application) PutHTTP(path string, handler http.Handler) {
	app.HandleHTTP(http.MethodPut, path, handler)
}

// Patch registers a route for HTTP PATCH requests.
func (app *Application) Patch(path string, next Next) {
	app.Handle(http.MethodPatch, path, next)
}

// PatchHTTP registers a standard net/http handler for HTTP PATCH requests.
func (app *Application) PatchHTTP(path string, handler http.Handler) {
	app.HandleHTTP(http.MethodPatch, path, handler)
}

// Delete registers a route for HTTP DELETE requests.
func (app *Application) Delete(path string, next Next) {
	app.Handle(http.MethodDelete, path, next)
}

// DeleteHTTP registers a standard net/http handler for HTTP DELETE requests.
func (app *Application) DeleteHTTP(path string, handler http.Handler) {
	app.HandleHTTP(http.MethodDelete, path, handler)
}

// Options registers a route for HTTP OPTIONS requests.
func (app *Application) Options(path string, next Next) {
	app.Handle(http.MethodOptions, path, next)
}

// OptionsHTTP registers a standard net/http handler for HTTP OPTIONS requests.
func (app *Application) OptionsHTTP(path string, handler http.Handler) {
	app.HandleHTTP(http.MethodOptions, path, handler)
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
		if idx := methodRootIndex(method); idx >= 0 {
			app.methodRoots[idx] = root
		}
		app.globalAllowed = app.allowed("*", "")
	}

	root.addRoute(path, next)
	if root.hasStaticParamSibling() {
		frozenRoot := root.freeze()
		if app.frozenTrees == nil {
			app.frozenTrees = make(map[string]*frozenNode)
		}
		app.frozenTrees[method] = frozenRoot
		if idx := methodRootIndex(method); idx >= 0 {
			app.frozenRoots[idx] = frozenRoot
		}
	}

	if pc := countParams(path); pc > app.maxParams {
		app.maxParams = pc
	}
}

// ServeFiles registers a GET route that serves files from root.
//
// The route path must end with /*filepath so the remaining request path can be
// captured and passed to http.FileServer. For example:
//
//	app.ServeFiles("/static/*filepath", http.Dir("./public"))
//
// A request to /static/css/app.css will serve /css/app.css from the provided
// file system. ServeFiles panics during registration if the pattern is invalid.
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

// ServeHTTP dispatches an HTTP request to the registered route tree.
//
// Application implements http.Handler through this method. ServeHTTP is
// responsible for route lookup, context pooling, middleware-executed handlers,
// response encoding, automatic OPTIONS handling, method-not-allowed handling,
// not-found handling, logging, and panic recovery.
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer app.recv(w, r)

	rel := r.URL.Path
	infoLogger := app.info
	errLogger := app.err

	if root, frozenRoot := app.rootsForMethod(r.Method); root != nil {
		if frozenRoot == nil {
			if next, params, _ := root.getValueFast(rel, app); next != nil {
				c := createCtx(app, w, r, params)
				val, err := next(c)
				userID := c.UserId()

				if err != nil {
					code, writeErr := app.handleError(c, err)
					app.putParams(params)
					releaseCtx(c)
					if writeErr != nil && errLogger != nil {
						errLogger.Print(formatRequestLog("error", "write_error", r, userID, rel, code, writeErr))
					}
					if errLogger != nil {
						errLogger.Print(formatRequestLog("error", "request", r, userID, rel, code, err))
					}

					return
				}

				if val != nil {
					code := c.statusCode
					if code == 0 {
						code = http.StatusOK
					}
					mt := c.responseMediaType()
					if !c.responseCommitted {
						writeCodeByMedia(w, mt, code)
					}
					err := c.writeMedia(mt, val)
					app.putParams(params)
					releaseCtx(c)
					if err != nil {
						if errLogger != nil {
							errLogger.Print(formatRequestLog("error", "write_error", r, userID, rel, code, err))
						}
						return
					}

					if infoLogger != nil {
						infoLogger.Print(formatRequestLog("info", "request", r, userID, rel, code, nil))
					}

					if rel, ok := val.(IRelease); ok {
						rel.Release()
					}
				} else {
					code := c.statusCode
					if code == 0 {
						code = http.StatusNoContent
					}
					committed := c.responseCommitted
					app.putParams(params)
					releaseCtx(c)
					if !committed {
						w.WriteHeader(code)
					}

					if infoLogger != nil {
						infoLogger.Print(formatRequestLog("info", "request", r, userID, rel, code, nil))
					}
				}

				return
			}
		} else {
			var match routeMatch
			frozenRoot.lookup(rel, app, &match)
			if match.callback != nil {

				c := createCtxWithRouteMatch(app, w, r, &match)
				val, err := match.callback(c)
				userID := c.UserId()

				if err != nil {
					code, writeErr := app.handleError(c, err)
					app.putParamValues(c.routeParamExtraValues)
					releaseCtx(c)
					if writeErr != nil && errLogger != nil {
						errLogger.Print(formatRequestLog("error", "write_error", r, userID, rel, code, writeErr))
					}
					if errLogger != nil {
						errLogger.Print(formatRequestLog("error", "request", r, userID, rel, code, err))
					}

					return
				}

				if val != nil {
					code := c.statusCode
					if code == 0 {
						code = http.StatusOK
					}
					mt := c.responseMediaType()
					if !c.responseCommitted {
						writeCodeByMedia(w, mt, code)
					}
					err := c.writeMedia(mt, val)
					app.putParamValues(c.routeParamExtraValues)
					releaseCtx(c)
					if err != nil {
						if errLogger != nil {
							errLogger.Print(formatRequestLog("error", "write_error", r, userID, rel, code, err))
						}
						return
					}

					if infoLogger != nil {
						infoLogger.Print(formatRequestLog("info", "request", r, userID, rel, code, nil))
					}

					if rel, ok := val.(IRelease); ok {
						rel.Release()
					}
				} else {
					code := c.statusCode
					if code == 0 {
						code = http.StatusNoContent
					}
					committed := c.responseCommitted
					app.putParamValues(c.routeParamExtraValues)
					releaseCtx(c)
					if !committed {
						w.WriteHeader(code)
					}

					if infoLogger != nil {
						infoLogger.Print(formatRequestLog("info", "request", r, userID, rel, code, nil))
					}
				}

				return
			}
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
	if e, ok := err.(*errFn); ok {
		code := e.code
		if cbErr := e.cb(c.w, c.r); cbErr == nil {
			return code, nil
		} else {
			err = cbErr
		}
	}

	if app.errorHandler != nil {
		if nextErr := app.errorHandler(c, err); nextErr == nil {
			return errCode(err), nil
		} else {
			err = nextErr
		}
	}

	code := errCode(err)
	mt := c.responseMediaType()
	writeCodeByMedia(c.w, mt, code)
	return code, c.writeMedia(mt, err.Error())
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

// Use appends middleware to this route group.
//
// Middleware added to a group affects only routes registered on that group after
// the call. It does not mutate the parent application or sibling groups.
func (g *RouteGroup) Use(middleware ...Middleware) {
	g.middleware = append(g.middleware, middleware...)
}

// Group creates a nested route group below the current group prefix.
//
// The child group inherits the current group's middleware and appends any
// middleware supplied to this call. The prefix must be empty or start with /.
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
//
// The final path is the group prefix joined with path. Middleware is composed in
// application, parent group, child group, then route order.
func (g *RouteGroup) Handle(method string, path string, next Next, middleware ...Middleware) {
	g.app.addRoute(method, joinPaths(g.prefix, path), wrapNext(next, g.app.middleware, g.middleware, Chain(middleware)))
}

// HandleHTTP registers a standard net/http handler on the group.
//
// This is the grouped equivalent of Application.HandleHTTP and is useful for
// mounting existing net/http handlers under an API prefix.
func (g *RouteGroup) HandleHTTP(method string, path string, handler http.Handler, middleware ...Middleware) {
	g.Handle(method, path, HTTPHandler(handler), middleware...)
}

// Get registers an HTTP GET route on the group.
func (g *RouteGroup) Get(path string, next Next) {
	g.Handle(http.MethodGet, path, next)
}

// GetHTTP registers a standard net/http handler for HTTP GET on the group.
func (g *RouteGroup) GetHTTP(path string, handler http.Handler) {
	g.HandleHTTP(http.MethodGet, path, handler)
}

// Head registers an HTTP HEAD route on the group.
func (g *RouteGroup) Head(path string, next Next) {
	g.Handle(http.MethodHead, path, next)
}

// HeadHTTP registers a standard net/http handler for HTTP HEAD on the group.
func (g *RouteGroup) HeadHTTP(path string, handler http.Handler) {
	g.HandleHTTP(http.MethodHead, path, handler)
}

// Post registers an HTTP POST route on the group.
func (g *RouteGroup) Post(path string, next Next) {
	g.Handle(http.MethodPost, path, next)
}

// PostHTTP registers a standard net/http handler for HTTP POST on the group.
func (g *RouteGroup) PostHTTP(path string, handler http.Handler) {
	g.HandleHTTP(http.MethodPost, path, handler)
}

// Put registers an HTTP PUT route on the group.
func (g *RouteGroup) Put(path string, next Next) {
	g.Handle(http.MethodPut, path, next)
}

// PutHTTP registers a standard net/http handler for HTTP PUT on the group.
func (g *RouteGroup) PutHTTP(path string, handler http.Handler) {
	g.HandleHTTP(http.MethodPut, path, handler)
}

// Patch registers an HTTP PATCH route on the group.
func (g *RouteGroup) Patch(path string, next Next) {
	g.Handle(http.MethodPatch, path, next)
}

// PatchHTTP registers a standard net/http handler for HTTP PATCH on the group.
func (g *RouteGroup) PatchHTTP(path string, handler http.Handler) {
	g.HandleHTTP(http.MethodPatch, path, handler)
}

// Delete registers an HTTP DELETE route on the group.
func (g *RouteGroup) Delete(path string, next Next) {
	g.Handle(http.MethodDelete, path, next)
}

// DeleteHTTP registers a standard net/http handler for HTTP DELETE on the group.
func (g *RouteGroup) DeleteHTTP(path string, handler http.Handler) {
	g.HandleHTTP(http.MethodDelete, path, handler)
}

// Options registers an HTTP OPTIONS route on the group.
func (g *RouteGroup) Options(path string, next Next) {
	g.Handle(http.MethodOptions, path, next)
}

// OptionsHTTP registers a standard net/http handler for HTTP OPTIONS on the group.
func (g *RouteGroup) OptionsHTTP(path string, handler http.Handler) {
	g.HandleHTTP(http.MethodOptions, path, handler)
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

			var cb Next
			if frozenRoot := app.frozenTrees[method]; frozenRoot != nil {
				cb, _, _ = frozenRoot.getValue(path, nil)
			} else {
				cb, _, _ = app.trees[method].getValueFast(path, nil)
			}
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

func (app *Application) rootsForMethod(method string) (*node, *frozenNode) {
	if idx := methodRootIndex(method); idx >= 0 {
		return app.methodRoots[idx], app.frozenRoots[idx]
	}
	return app.trees[method], app.frozenTrees[method]
}

// ListenAndServe starts an HTTP server on network and addr.
//
// The application is installed as the root handler. Optional functions receive
// the created *http.Server before Serve is called, which lets callers configure
// timeouts, TLSNextProto, ErrorLog, and other server fields.
func (app *Application) ListenAndServe(network string, addr string, fns ...func(*http.Server)) error {

	l, err := net.Listen(network, addr)

	if err != nil {
		return err
	}

	defer l.Close()

	return app.serve(l, fns...)
}

// ListenAndServeTLS starts an HTTPS server on network and addr.
//
// tlsConfig is passed to tls.Listen. Optional functions receive the created
// *http.Server before Serve is called, matching ListenAndServe.
func (app *Application) ListenAndServeTLS(network string, addr string, tlsConfig *tls.Config, fns ...func(*http.Server)) error {

	l, err := tls.Listen(network, addr, tlsConfig)

	if err != nil {
		return err
	}

	defer l.Close()

	return app.serve(l, fns...)
}

// Shutdown gracefully shuts down the HTTP server created by ListenAndServe or
// ListenAndServeTLS.
//
// The context controls the shutdown deadline. If the application has not started
// a server through this package, Shutdown returns ErrServerNotInitialized.
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

// Inspect returns a textual description of the application.
//
// The method is reserved for future diagnostics and currently returns an empty
// string.
func (app *Application) Inspect() string {
	return ""
}

// Logf writes a formatted message to the informational logger when one is set.
//
// Calls are ignored when no informational logger has been configured.
func (app *Application) Logf(format string, v ...any) {
	if app.info != nil {
		app.info.Printf(format, v...)
	}
}

// Errf writes a formatted message to the error logger when one is set.
//
// Calls are ignored when no error logger has been configured.
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

func (app *Application) getParamValues() *[]string {
	values := app.paramValuesPool.Get().(*[]string)
	if uint16(cap(*values)) < app.maxParams {
		*values = make([]string, 0, app.maxParams)
	} else {
		*values = (*values)[:0]
	}
	return values
}

func (app *Application) putParamValues(values *[]string) {
	if values != nil {
		app.paramValuesPool.Put(values)
	}
}

func (app *Application) recv(w http.ResponseWriter, r *http.Request) {
	if rcv := recover(); rcv != nil {
		writeCode(w, r, http.StatusInternalServerError)
		if app.panic != nil {
			app.panic(w, r, rcv)
		} else {
			app.Errf("%s", formatPanicLog(r, rcv))
		}
	}
}

func formatRequestLog(level string, event string, r *http.Request, userID uint64, path string, status int, err error) string {
	var b strings.Builder
	b.Grow(160)
	appendLogfmtString(&b, "level", level)
	appendLogfmtString(&b, "event", event)
	appendLogfmtString(&b, "remote_addr", r.RemoteAddr)
	appendLogfmtString(&b, "host", r.Host)
	appendLogfmtUint(&b, "user_id", userID)
	appendLogfmtString(&b, "method", r.Method)
	appendLogfmtString(&b, "path", path)
	appendLogfmtInt(&b, "status", status)
	if err != nil {
		appendLogfmtString(&b, "error", err.Error())
	}
	return b.String()
}

func formatPanicLog(r *http.Request, recovered any) string {
	var b strings.Builder
	b.Grow(128)
	appendLogfmtString(&b, "level", "error")
	appendLogfmtString(&b, "event", "panic")
	appendLogfmtString(&b, "remote_addr", r.RemoteAddr)
	appendLogfmtString(&b, "host", r.Host)
	appendLogfmtString(&b, "method", r.Method)
	appendLogfmtString(&b, "path", r.URL.Path)
	appendLogfmtString(&b, "error", toString(recovered))
	return b.String()
}

func appendLogfmtString(b *strings.Builder, key string, value string) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	b.WriteString(strconv.Quote(value))
}

func appendLogfmtInt(b *strings.Builder, key string, value int) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	b.WriteString(strconv.Itoa(value))
}

func appendLogfmtUint(b *strings.Builder, key string, value uint64) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	b.WriteString(strconv.FormatUint(value, 10))
}

func toString(v any) string {
	return fmt.Sprint(v)
}
