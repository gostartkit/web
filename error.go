package web

import (
	"errors"
)

var (
	// ErrUnauthorized 401
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden 403
	ErrForbidden = errors.New("forbidden")
	// ErrUnExpected unexpected error
	ErrUnExpected = errors.New("unexpected")
	// ErrNotFound
	ErrNotFound = errors.New("not found")

	// ErrContentTypeInvalid content-type not supported
	ErrContentTypeInvalid = errors.New("content-type not supported")
	// ErrMethodNotImplemented method not implemented
	ErrMethodNotImplemented = errors.New("method not implemented")
)
