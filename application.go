package web

import (
	"net/http"
	"sync"
)

var app *Application
var once sync.Once

// Application is type of a web.Application
type Application struct {
	Context *Context
	Keys    []string
}

// CreateApplication return a singleton web.Application
func CreateApplication() *Application {
	once.Do(func() {
		app = createApplication()
	})
	return app
}

// Use add the given middleware function to web.Application.
func (app *Application) Use(fn func(ctx *Context)) *Application {
	return app
}

// On add event
func (app *Application) On(name string, fn func(err string, ctx *Context)) *Application {
	return app
}

// Listen addr
func (app *Application) Listen(addr string) *Application {
	http.ListenAndServe(addr, nil)
	return app
}

// Inspect method
func (app *Application) Inspect() string {
	return ""
}

// createApplication return a web.Application
func createApplication() *Application {
	return &Application{}
}

// createContext return a web.Context
func createContext() *Context {
	return &Context{}
}
