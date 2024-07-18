package web

import (
	"errors"
)

var (
	// ErrMovedPermanently 301 moved permanently
	ErrMovedPermanently = errors.New("moved permanently")
	// ErrFound 302 found
	ErrFound = errors.New("found")
	// ErrTemporaryRedirect 307 temporary redirect
	ErrTemporaryRedirect = errors.New("temporary redirect")
	// ErrPermanentRedirect 308 permanent redirect
	ErrPermanentRedirect = errors.New("permanent redirect")
	// ErrUnauthorized 401 unauthorized
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden 403 forbidden
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound 404 not found
	ErrNotFound = errors.New("not found")
	// ErrMethodNotAllowed 405 method not allowed
	ErrMethodNotAllowed = errors.New("method not allowed")
	// ErrNotImplemented 501 not implemented
	ErrNotImplemented = errors.New("not implemented")
	// ErrContentType content-type not supported
	ErrContentType = errors.New("content-type not supported")
	// ErrCors cross origin request blocked
	ErrCors = errors.New("cross origin request blocked")
	// ErrCallBack callback
	ErrCallBack = errors.New("callback")
	// ErrUnExpected unexpected error
	ErrUnExpected = errors.New("unexpected")

	// ErrNotVerified object not verified
	ErrNotVerified = errors.New("object not verified")
	// ErrInvalid object invalid
	ErrInvalid = errors.New("object invalid")
)
