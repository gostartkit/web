package web

import (
	"errors"
)

var (
	// ErrMovedPermanently 301
	ErrMovedPermanently = errors.New("moved permanently")
	// ErrFound 302
	ErrFound = errors.New("found")
	// ErrTemporaryRedirect 307
	ErrTemporaryRedirect = errors.New("temporary redirect")
	// ErrPermanentRedirect 308
	ErrPermanentRedirect = errors.New("permanent redirect")
	// ErrUnauthorized 401
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden 403
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound 404
	ErrNotFound = errors.New("not found")
	// ErrUnExpected unexpected error
	ErrUnExpected = errors.New("unexpected")
	// ErrContentType content-type not supported
	ErrContentType = errors.New("content-type not supported")
	// ErrMethodNotImplemented method not implemented
	ErrMethodNotImplemented = errors.New("method not implemented")
)
