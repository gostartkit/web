package web

import "net/http"

// Any Any
type Any interface{}

// Callback function
type Callback func(c *Ctx) (Any, error)

// PanicCallback function
type PanicCallback func(http.ResponseWriter, *http.Request, Any)

// Middleware
type Middleware func(Callback) Callback

// Chain middleware chain
type Chain []Middleware

// Reader function
type Reader func(c *Ctx, v Any) error

// Writer function
type Writer func(c *Ctx, v Any) error
