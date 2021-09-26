package web

import (
	"errors"
	"io"
)

// binaryReader decode data from binary
func binaryReader(ctx *Context, v Data) error {
	if app().binaryReader != nil {
		return app().binaryReader(ctx, v)
	}
	return errors.New("binaryReader not implemented")
}

// binaryWriter encode data to binary
func binaryWriter(w io.Writer, v Data) error {
	if app().binaryWriter != nil {
		return app().binaryWriter(w, v)
	}
	return errors.New("binaryWriter not implemented")
}
