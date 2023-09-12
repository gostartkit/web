package web

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	_app  *Application
	_once sync.Once
)

// Data Data
type Data interface{}

// Callback function
type Callback func(c *WebContext) (Data, error)

// PanicCallback function
type PanicCallback func(http.ResponseWriter, *http.Request, interface{})

// Application is type of a web.Application
type Application struct {
	trees          map[string]*node
	formReader     Reader
	formDataReader Reader
	binaryReader   Reader
	binaryWriter   Writer
	logger         *log.Logger
	panic          PanicCallback
	paramsPool     sync.Pool
	maxParams      uint16
	extension      string

	NotFound http.Handler
}

// CreateApplication return a singleton web.Application
func CreateApplication() *Application {
	_once.Do(func() {
		_app = &Application{}
	})
	return _app
}

// SetFormReader set formReader
func (app *Application) SetFormReader(formReader Reader) {
	app.formReader = formReader
}

// SetFormDataReader set formDataReader
func (app *Application) SetFormDataReader(formDataReader Reader) {
	app.formDataReader = formDataReader
}

// SetBinaryReader set binaryReader
func (app *Application) SetBinaryReader(binaryReader Reader) {
	app.binaryReader = binaryReader
}

// SetBinaryWriter set binaryWriter
func (app *Application) SetBinaryWriter(binaryWriter Writer) {
	app.binaryWriter = binaryWriter
}

// SetLogger set Logger
func (app *Application) SetLogger(logger *log.Logger) {
	app.logger = logger
}

// SetPanic set set Panic
func (app *Application) SetPanic(panic PanicCallback) {
	app.panic = panic
}

// SetExtension set Extension
func (app *Application) SetExtension(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	app.extension = ext
}

// Use Add the given callback function to this application.middlewares.
func (app *Application) Use(path string, cb Callback) {

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
}

// ServeFiles ("/src/*filepath", http.Dir("/var/www"))
func (app *Application) ServeFiles(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	app.Get(path, func(c *WebContext) (Data, error) {
		c.r.URL.Path = c.Param("filepath")
		fileServer.ServeHTTP(c.w, c.r)
		return nil, nil
	})
}

// ServeHTTP w, r
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer app.recv(w, r)

	path := r.URL.Path

	if filepath.Ext(path) != app.extension {
		http.NotFound(w, r)
		return
	}

	if root := app.trees[r.Method]; root != nil {

		if callback, params, _ := root.getValue(path, app.getParams); callback != nil {

			c := createWebContext(w, r, params)

			val, err := callback(c)

			app.putParams(params)

			if err != nil {

				status := http.StatusBadRequest

				switch err {
				case ErrUnauthorized:
					status = http.StatusUnauthorized
				case ErrForbidden:
					status = http.StatusForbidden
				}

				w.WriteHeader(status)
				c.Write(err.Error())

				app.logf("%s %s %d %s %s %d %v", r.RemoteAddr, r.Host, c.UserID(), r.Method, path, status, err)

				return
			}

			if val != nil {

				status := http.StatusOK

				switch r.Method {
				case http.MethodPost:
					status = http.StatusCreated
				}

				w.WriteHeader(status)
				c.Write(val)

				return

			}

			w.WriteHeader(http.StatusNoContent)

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

func (app *Application) serve(listener net.Listener, fns ...func(*http.Server)) error {

	mux := http.NewServeMux()

	mux.Handle("/", app)

	srv := &http.Server{
		Handler: mux,
	}

	for _, fn := range fns {
		fn(srv)
	}

	if app.paramsPool.New == nil && app.maxParams > 0 {
		app.paramsPool.New = func() interface{} {
			ps := make(Params, 0, app.maxParams)
			return &ps
		}
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
		w.WriteHeader(http.StatusInternalServerError)
		if app.panic != nil {
			app.panic(w, r, rcv)
		} else {
			app.logf("%s %s %s %s rcv: %v", r.RemoteAddr, r.Host, r.Method, r.URL.Path, rcv)
		}
	}
}
