package web

import (
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var _app *Application
var _once sync.Once

// Callback function
type Callback func(ctx *Context)

// Application is type of a web.Application
type Application struct {
	middlewares []Callback
	logger      *log.Logger
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
		middlewares: []Callback{},
		logger:      log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}

	return app
}

// Use Add the given callback function to this application.middlewares.
func (app *Application) Use(callback Callback) {
	app.middlewares = append(app.middlewares, callback)
}

// Resource Add the given callback function to this application.middlewares.
func (app *Application) Resource(path string, callback Callback) {
	app.middlewares = append(app.middlewares, callback)
}

// On add event
func (app *Application) On(name string, callback Callback) {

}

// ServeHTTP
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	ctx := newContext(w, r)

	for i := range app.middlewares {
		callback := app.middlewares[i]
		callback(ctx)
	}

	endTime := time.Now()

	app.logger.Printf("%s", endTime.Sub(startTime))
}

// ListenAndServe on addr
func (app *Application) ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal("Listen:", err)
	}

	app.logger.Printf("web.go serving %s\n", l.Addr())

	return http.Serve(l, app)
}

// ListenAndServeTLS on addr
func (app *Application) ListenAndServeTLS(addr, certFile, keyFile string) error {
	l, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal("Listen:", err)
	}

	app.logger.Printf("web.go serving %s\n", l.Addr())

	return http.ServeTLS(l, app, certFile, keyFile)
}

// Inspect method
func (app *Application) Inspect() string {
	return ""
}

// Log method
func (app *Application) Log(line string) {

}

// newContext return a web.Context
func newContext(w http.ResponseWriter, r *http.Request) *Context {

	ctx := &Context{
		Response: &Response{
			w: w,
		},
		Request: &Request{
			r: r,
		},
	}

	return ctx
}
