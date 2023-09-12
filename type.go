package web

// Reader function
type Reader func(ctx *WebContext, v Data) error

// Writer function
type Writer func(ctx *WebContext, v Data) error
