package web

// Reader function
type Reader func(ctx *Context, v Data) error

// Writer function
type Writer func(ctx *Context, v Data) error
