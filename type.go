package web

import "net/http"

// Any Any
type Any interface{}

// Callback function
type Callback func(c *WebContext) (Any, error)

// PanicCallback function
type PanicCallback func(http.ResponseWriter, *http.Request, interface{})

// Middleware
type Middleware func(Callback) Callback

// Chain middleware chain
type Chain []Middleware

// Reader function
type Reader func(c *WebContext, v Any) error

// Writer function
type Writer func(c *WebContext, v Any) error
