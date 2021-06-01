package web

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/template"
)

var (
	_app  *Application
	_once sync.Once
)

// Data Data
type Data interface{}

// Callback function
type Callback func(ctx *Context) (Data, error)

// PanicCallback function
type PanicCallback func(http.ResponseWriter, *http.Request, interface{})

// Application is type of a web.Application
type Application struct {
	trees      map[string]*node
	logger     *log.Logger
	panic      PanicCallback
	paramsPool sync.Pool
	maxParams  uint16
	template   *template.Template

	NotFound http.Handler
}

// Create return a singleton web.Application
func Create() *Application {
	_once.Do(func() {
		_app = &Application{}
	})
	return _app
}

// SetLogger set Logger
func (app *Application) SetLogger(logger *log.Logger) {
	app.logger = logger
}

// SetPanic set Logger
func (app *Application) SetPanic(panic PanicCallback) {
	app.panic = panic
}

// SetTemplate set template
func (app *Application) SetTemplate(template *template.Template) {
	app.template = template
}

// Template get template
func (app *Application) Template() *template.Template {
	if app.template == nil {
		app.template = template.New("TOP")
	}
	return app.template
}

// Execute execute template
func (app *Application) Execute(wr io.Writer, val interface{}) error {
	return app.Template().Execute(wr, val)
}

// ExecuteTemplate execute template by name
func (app *Application) ExecuteTemplate(wr io.Writer, name string, val interface{}) error {
	return app.Template().ExecuteTemplate(wr, name, val)
}

// Use Add the given callback function to this application.middlewares.
func (app *Application) Use(path string, callback Callback) {

}

// On add event
func (app *Application) On(name string, cb Callback) {

}

// Get method
func (app *Application) Get(path string, cb Callback) {
	app.addRoute(http.MethodGet, path, cb)
}

// Head method
func (app *Application) Head(path string, cb Callback) {
	app.addRoute(http.MethodHead, path, cb)
}

// Post method
func (app *Application) Post(path string, cb Callback) {
	app.addRoute(http.MethodPost, path, cb)
}

// Put method
func (app *Application) Put(path string, cb Callback) {
	app.addRoute(http.MethodPut, path, cb)
}

// Patch method
func (app *Application) Patch(path string, cb Callback) {
	app.addRoute(http.MethodPatch, path, cb)
}

// Delete method
func (app *Application) Delete(path string, cb Callback) {
	app.addRoute(http.MethodDelete, path, cb)
}

// Options method
func (app *Application) Options(path string, cb Callback) {
	app.addRoute(http.MethodOptions, path, cb)
}

func (app *Application) addRoute(method, path string, cb Callback) {

	if method == "" {
		panic("method must not be empty")
	}

	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if cb == nil {
		panic("callback must not be nil")
	}

	if app.trees == nil {
		app.trees = make(map[string]*node)
	}

	root := app.trees[method]

	if root == nil {
		root = new(node)
		app.trees[method] = root
	}

	root.addRoute(path, cb)

	if pc := countParams(path); pc > app.maxParams {
		app.maxParams = pc
	}

	if app.paramsPool.New == nil && app.maxParams > 0 {
		app.paramsPool.New = func() interface{} {
			ps := make(Params, 0, app.maxParams)
			return &ps
		}
	}
}

// ServeFiles ("/src/*filepath", http.Dir("/var/www"))
func (app *Application) ServeFiles(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	app.Get(path, func(ctx *Context) (Data, error) {
		ctx.Request.URL.Path = ctx.Param("filepath")
		fileServer.ServeHTTP(ctx.ResponseWriter, ctx.Request)
		return nil, nil
	})
}

// ServeHTTP w, r
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer app.recv(w, r)
	path := r.URL.Path

	if root := app.trees[r.Method]; root != nil {

		if callback, params, _ := root.getValue(path, app.getParams); callback != nil {

			ctx := createContext(w, r, params)
			app.putParams(params)

			val, err := callback(ctx)

			if err != nil {

				switch err {
				case ErrUnauthorized:
					ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
				case ErrForbidden:
					ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
				default:
					ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
				}

				app.logf("%s %s %d %s %s %s", r.RemoteAddr, r.Host, ctx.UserID(), r.Method, path, err)

				if err := ctx.Write(err.Error()); err != nil {
					app.logf("%s %s %d %s %s %s", r.RemoteAddr, r.Host, ctx.UserID(), r.Method, path, err)
				}

				return
			}

			if val != nil {
				if strings.HasSuffix(ctx.ContentType(), "html") {
					viewName := ctx.Query("$viewName")
					if viewName == "" {
						viewName = replace(path, '/', '_')
					}
					if err := app.ExecuteTemplate(w, viewName, val); err != nil {
						app.logf("%s %s %d %s %s %s", r.RemoteAddr, r.Host, ctx.UserID(), r.Method, path, err)
					}
				} else {
					if err := ctx.Write(val); err != nil {
						app.logf("%s %s %d %s %s %s", r.RemoteAddr, r.Host, ctx.UserID(), r.Method, path, err)
					}
				}
			}

			return
		}
	}

	if app.NotFound != nil {
		app.NotFound.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

// ListenAndServe Serve with options on addr
func (app *Application) ListenAndServe(addr string, fns ...func(*http.Server)) error {

	l, err := net.Listen("tcp", addr)

	if err != nil {
		return err
	}

	defer l.Close()

	return app.serve(addr, l, fns...)
}

// ListenAndServeTLS Serve with tls and options on addr
func (app *Application) ListenAndServeTLS(addr string, tlsConfig *tls.Config, fns ...func(*http.Server)) error {

	l, err := tls.Listen("tcp", addr, tlsConfig)

	if err != nil {
		return err
	}

	defer l.Close()

	return app.serve(addr, l, fns...)
}

func (app *Application) serve(addr string, listener net.Listener, fns ...func(*http.Server)) error {

	mux := http.NewServeMux()

	mux.Handle("/", app)

	srv := &http.Server{
		Handler: mux,
	}

	for _, fn := range fns {
		fn(srv)
	}

	if err := srv.Serve(listener); err != nil {
		return err
	}

	if err := srv.Close(); err != nil {
		return err
	}

	return nil
}

// Inspect method
func (app *Application) Inspect() string {
	return ""
}

// logf write log
func (app *Application) logf(format string, v ...interface{}) {
	if app.logger == nil {
		app.logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	}

	app.logger.Printf(format, v...)
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
		if app.panic != nil {
			app.panic(w, r, rcv)
		} else {
			app.logf("%s %s %s: %v", r.Host, r.Method, r.URL.Path, rcv)
		}
	}
}
