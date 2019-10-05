package web

import (
	"net/http"
	"sync"
)

var _app *Application
var _once sync.Once

// Callback function
type Callback func(ctx *Context)

// Application is type of a web.Application
type Application struct {
	middlewares []Callback
	keys        []string
}

// Singleton return a singleton web.Application
func Singleton() *Application {
	_once.Do(func() {
		_app = newApplication()
	})
	return _app
}

// newApplication return a web.Application
func newApplication() *Application {
	app := &Application{}

	return app
}

// Use add the given middleware function to web.Application.
func (app *Application) Use(callback Callback) {

}

// On add event
func (app *Application) On(name string, callback Callback) {

}

// Listen addr
func (app *Application) Listen(addr string) {
	http.ListenAndServe(addr, nil)
}

// Inspect method
func (app *Application) Inspect() string {
	return ""
}

// newContext return a web.Context
func newContext() *Context {
	return &Context{}
}
