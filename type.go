package web

import "io"

// Reader function
type Reader func(ctx *Context, v Data) error

// Writer function
type Writer func(w io.Writer, v Data) error

// ViewWriter function
type ViewWriter func(w io.Writer, ctx *Context, v Data) error
