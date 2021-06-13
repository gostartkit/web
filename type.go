package web

import "io"

// Reader function
type Reader func(r io.ReadCloser, v interface{}) error

// Writer function
type Writer func(io.Writer, interface{}) error

// HtmlWriter function
type HtmlWriter func(io.Writer, *Context, interface{}) error
