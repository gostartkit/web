package web

import (
	"errors"
	"io"
)

var (
	_binaryReader Reader
	_binaryWriter Writer
)

// SetBinaryReader set binaryReader
func SetBinaryReader(r Reader) {
	_binaryReader = r
}

// SetBinaryWriter set binaryWriter
func SetBinaryWriter(w Writer) {
	_binaryWriter = w
}

// binaryReader decode data from binary
func binaryReader(r io.ReadCloser, v interface{}) error {
	if _binaryReader != nil {
		return _binaryReader(r, v)
	}
	return errors.New("binaryReader not implemented")
}

// binaryWriter encode data to binary
func binaryWriter(w io.Writer, v interface{}) error {
	if _binaryWriter != nil {
		return _binaryWriter(w, v)
	}
	return errors.New("binaryWriter not implemented")
}
