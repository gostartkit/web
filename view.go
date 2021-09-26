package web

import (
	"errors"
	"io"
)

// viewWriter encode data to html
func viewWriter(w io.Writer, ctx *Context, v interface{}) error {
	if app().viewWriter != nil {
		return app().viewWriter(w, ctx, v)
	}
	return errors.New("viewWriter not implemented")
}
