package web

import "io"

// Reader function
type Reader func(ctx *Context, v Data) error

// Writer function
type Writer func(w io.Writer, v Data) error

// HtmlWriter function
type HtmlWriter func(w io.Writer, ctx *Context, v Data) error
