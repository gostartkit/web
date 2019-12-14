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
	trees       map[string]*node
	middlewares Middlewares
	logger      *log.Logger
	panic       PanicCallback
	paramsPool  sync.Pool
	maxParams   uint16
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
		// middlewares: middlewares{},
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
	m := Middleware{
		Path:     path,
		Callback: callback,
	}
	app.middlewares = append(app.middlewares, m)
}

// Resource map controller path
func (app *Application) Resource(path string, controller Controller) {

	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}

	pos := len(path) - 1

	if pos >= 0 {
		if path[pos] != '/' {
			path = path + "/"
		}
	} else {
		path = "/"
	}

	app.Get(path, controller.Index)
	app.Post(path, controller.Create)
	app.Get(path+":id", controller.Detail)
	app.Patch(path+":id", controller.Update)
	app.Put(path+":id", controller.Update)
	app.Delete(path+":id", controller.Destroy)
}

// On add event
func (app *Application) On(name string, callback Callback) {

}

// Get method
func (app *Application) Get(path string, callback Callback) {
	app.addRoute(http.MethodGet, path, callback)
}

// Post method
func (app *Application) Post(path string, callback Callback) {
	app.addRoute(http.MethodPost, path, callback)
}

// Put method
func (app *Application) Put(path string, callback Callback) {
	app.addRoute(http.MethodPut, path, callback)
}

// Patch method
func (app *Application) Patch(path string, callback Callback) {
	app.addRoute(http.MethodPatch, path, callback)
}

// Delete method
func (app *Application) Delete(path string, callback Callback) {
	app.addRoute(http.MethodDelete, path, callback)
}

// Options method
func (app *Application) Options(path string, callback Callback) {
	app.addRoute(http.MethodOptions, path, callback)
}

func (app *Application) addRoute(method, path string, callback Callback) {

	if method == "" {
		panic("method must not be empty")
	}

	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if callback == nil {
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

	root.addRoute(path, callback)

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

// ServeHTTP
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer app.recv(w, r)
	path := r.URL.Path

	if root := app.trees[r.Method]; root != nil {

		if callback, params, tsr := root.getValue(path, app.getParams); callback != nil {

			ctx := newContext(w, r, params)
			app.putParams(params)

			app.middlewares.exec(path, ctx)

			runTime := time.Now()
			callback(ctx)
			endTime := time.Now()

			app.logf("%s %s %s %s %s", r.Method, path, endTime.Sub(startTime), runTime.Sub(startTime), endTime.Sub(runTime))

			return

		} else if r.Method != http.MethodConnect && path != "/" {

			code := http.StatusMovedPermanently

			if r.Method != http.MethodGet {
				code = http.StatusPermanentRedirect
			}

			if tsr {

				if len(path) > 1 && path[len(path)-1] == '/' {
					r.URL.Path = path[:len(path)-1]
				} else {
					r.URL.Path = path + "/"
				}

				http.Redirect(w, r, r.URL.String(), code)

				return
			}
		}
	}

	http.NotFound(w, r)
}

// ListenAndServe Serve with options on addr
func (app *Application) ListenAndServe(config *ServerConfig, options ...func(*http.Server)) error {

	l, err := net.Listen("tcp", config.Addr)

	if err != nil {
		log.Fatal("Listen:", err)
	}

	defer l.Close()

	return app.serve(config, l, options...)
}

// ListenAndServeTLS Serve with tls and options on addr
func (app *Application) ListenAndServeTLS(config *ServerConfig, tlsConfig *tls.Config, options ...func(*http.Server)) error {

	l, err := tls.Listen("tcp", config.Addr, tlsConfig)

	if err != nil {
		log.Fatal("Listen:", err)
	}

	defer l.Close()

	return app.serve(config, l, options...)
}

func (app *Application) serve(config *ServerConfig, listener net.Listener, options ...func(*http.Server)) error {

	mux := http.NewServeMux()

	mux.Handle("/", app)

	srv := &http.Server{
		Handler:           mux,
		ReadTimeout:       config.ReadTimeout * time.Second,
		ReadHeaderTimeout: config.ReadHeaderTimeout * time.Second,
		WriteTimeout:      config.WriteTimeout * time.Second,
		IdleTimeout:       config.IdleTimeout * time.Second,
	}

	for _, option := range options {
		option(srv)
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
			app.logf("web.go: %v", rcv)
		}
	}
}
