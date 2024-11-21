package web

import (
	"errors"
)

var (
	// ErrMovedPermanently 301 moved permanently
	ErrMovedPermanently = errors.New("301 moved permanently")
	// ErrFound 302 found
	ErrFound = errors.New("302 found")
	// ErrTemporaryRedirect 307 temporary redirect
	ErrTemporaryRedirect = errors.New("307 temporary redirect")
	// ErrPermanentRedirect 308 permanent redirect
	ErrPermanentRedirect = errors.New("308 permanent redirect")
	// ErrBadRequest 400 bad request
	ErrBadRequest = errors.New("400 bad request")
	// ErrUnauthorized 401 unauthorized
	ErrUnauthorized = errors.New("401 unauthorized")
	// ErrForbidden 403 forbidden
	ErrForbidden = errors.New("403 forbidden")
	// ErrNotFound 404 not found
	ErrNotFound = errors.New("404 not found")
	// ErrMethodNotAllowed 405 method not allowed
	ErrMethodNotAllowed = errors.New("405 method not allowed")
	// ErrNotImplemented 501 not implemented
	ErrNotImplemented = errors.New("501 not implemented")
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
