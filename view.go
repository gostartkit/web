package web

import (
	"io"
)

// viewWriter encode data to html
func viewWriter(w io.Writer, v interface{}) error {
	if app().viewWriter != nil {
		return app().viewWriter(w, v)
	}
	return ErrViewWriterNotImplemented
}
