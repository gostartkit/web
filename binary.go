package web

import (
	"errors"
	"io"
)

// binaryReader decode data from binary
func binaryReader(ctx *Context, v Data) error {
	if App().binaryReader != nil {
		return App().binaryReader(ctx, v)
	}
	return errors.New("binaryReader not implemented")
}

// binaryWriter encode data to binary
func binaryWriter(w io.Writer, v Data) error {
	if App().binaryWriter != nil {
		return App().binaryWriter(w, v)
	}
	return errors.New("binaryWriter not implemented")
}
