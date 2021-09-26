package web

import "errors"

var (
	// ErrUnauthorized 401
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden 403
	ErrForbidden = errors.New("forbidden")

	// ErrViewWriterNotImplemented viewWriter not implemented
	ErrViewWriterNotImplemented = errors.New("viewWriter not implemented")
	// ErrFormDataReaderNotImplemented formDataReader not implemented
	ErrFormDataReaderNotImplemented = errors.New("formDataReader not implemented")
	// ErrBinaryReaderNotImplemented binaryReader not implemented
	ErrBinaryReaderNotImplemented = errors.New("binaryReader not implemented")
	// ErrBinaryWriterNotImplemented binaryWriter not implemented
	ErrBinaryWriterNotImplemented = errors.New("binaryWriter not implemented")
)
