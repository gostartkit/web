package web

import (
	"errors"
	"io"
)

var (
	_htmlWriter HtmlWriter
)

// SetHtmlWriter set htmlWriter
func SetHtmlWriter(w HtmlWriter) {
	_htmlWriter = w
}

// htmlWriter encode data to html
func htmlWriter(w io.Writer, ctx *Context, v interface{}) error {
	if _htmlWriter != nil {
		return _htmlWriter(w, ctx, v)
	}
	return errors.New("htmlWriter not implemented")
}
