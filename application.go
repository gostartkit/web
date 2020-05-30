package web

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	_app  *Application
	_once sync.Once
)

// Callback function
type Callback func(ctx *Context)

// PanicCallback function
type PanicCallback func(http.ResponseWriter, *http.Request, interface{})

// Application is type of a web.Application
type Application struct {
	trees      map[string]*node
	logger     *log.Logger
	panic      PanicCallback
	paramsPool sync.Pool
	maxParams  uint16
	NotFound   http.Handler
}

// Create return a singleton web.Application
func Create() *Application {
	_once.Do(func() {
		_app = newApplication()
	})
	return _app
}

// newApplication return a web.Application
func newApplication() *Application {
	app := &Application{
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
	return app
}

// SetLogger set Logger
func (app *Application) SetLogger(logger *log.Logger) {
	app.logger = logger
}

// SetPanic set Logger
func (app *Application) SetPanic(panic PanicCallback) {
	app.panic = panic
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

	app.Get(path, func(ctx *Context) {
		ctx.Request.URL.Path = ctx.Param("filepath")
		fileServer.ServeHTTP(ctx.ResponseWriter, ctx.Request)
	})
}

// ServeHTTP w, r
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer app.recv(w, r)
	path := r.URL.Path

	if root := app.trees[r.Method]; root != nil {

		if callback, params, _ := root.getValue(path, app.getParams); callback != nil {

			ctx := createContext(w, r, params)
			app.putParams(params)

			runTime := time.Now()

			callback(ctx)

			endTime := time.Now()

			app.logf("%s %s %s %s %s %s", r.Host, r.Method, path, endTime.Sub(startTime), runTime.Sub(startTime), endTime.Sub(runTime))

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

	defer func() {
		err := srv.Close()
		if err != nil {
			app.logf("srv: %v", err)
		}
	}()

	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			app.logf("web.go: %v", err)
		}

		close(idleConnsClosed)
	}()

	app.logf("web.go(%d) %s", os.Getpid(), listener.Addr())

	if err := srv.Serve(listener); err != nil {
		app.logf("web.go: %v", err)
	}

	<-idleConnsClosed

	return errors.New("web.go: exit")
}

// Inspect method
func (app *Application) Inspect() string {
	return ""
}

func (app *Application) logf(format string, v ...interface{}) {

	if app.logger != nil {
		app.logger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
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
		if app.panic != nil {
			app.panic(w, r, rcv)
		} else {
			app.logf("%s %s %s: %v", r.Host, r.Method, r.URL.Path, rcv)
		}
	}
}
