package web

import "sync"

var app *Application
var once sync.Once

// Application is type of a web.Application
type Application struct {
	ctx *Context
}

// CreateApplication return a singleton web.Application
func CreateApplication() *Application {
	once.Do(func() {
		app = createApplication()
	})
	return app
}

// createApplication return a web.Application
func createApplication() *Application {
	return &Application{}
}

// Use add the given middleware function to web.Application.
func (app *Application) Use(fn func(ctx *Context)) *Application {
	fn(app.ctx)
	return app
}
