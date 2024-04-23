package web

import "errors"

var (
	// ErrUnauthorized 401
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden 403
	ErrForbidden = errors.New("forbidden")
	// ErrUnExpected unexpected error
	ErrUnExpected = errors.New("unexpected")

	// ErrContentTypeNotSupported content-type not supported
	ErrContentTypeNotSupported = errors.New("content-type not supported")
	// ErrMethodNotImplemented method not implemented
	ErrMethodNotImplemented = errors.New("method not implemented")
)
