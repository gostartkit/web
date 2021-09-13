package web

import (
	"errors"
	"io"
)

// viewWriter encode data to html
func viewWriter(w io.Writer, ctx *Context, v interface{}) error {
	if App().viewWriter != nil {
		return App().viewWriter(w, ctx, v)
	}
	return errors.New("htmlWriter not implemented")
}
