package web

import (
	"errors"
)

var (
	// ErrMovedPermanently represents an HTTP 301 Moved Permanently error.
	// This error indicates that the requested resource has been permanently moved to a new URL.
	// Typically used in web applications when a redirect to a new permanent location is required.
	// Example usage: return this error when a webpage or API endpoint has been relocated permanently,
	// and the client should update its bookmarks or references accordingly.
	ErrMovedPermanently = errors.New("moved permanently")

	// ErrFound represents an HTTP 302 Found error.
	// This error signifies that the requested resource has been temporarily moved to a different URL.
	// Commonly used in web redirection scenarios where the resource is available at a different location
	// for the time being, but the original URL might still be valid in the future.
	// Example: a temporary redirect during site maintenance.
	ErrFound = errors.New("found")

	// ErrTemporaryRedirect represents an HTTP 307 Temporary Redirect error.
	// This error indicates a temporary redirection where the client should retry the request
	// at a different URL while preserving the original request method (e.g., POST remains POST).
	// Useful in scenarios like load balancing or temporary resource unavailability.
	// Note: Unlike 302, 307 explicitly requires maintaining the HTTP method.
	ErrTemporaryRedirect = errors.New("temporary redirect")

	// ErrPermanentRedirect represents an HTTP 308 Permanent Redirect error.
	// This error denotes a permanent redirection to a new URL, requiring the client to update its
	// references while preserving the original request method.
	// Suitable for permanent resource relocation where the method (e.g., POST, GET) must remain unchanged.
	// Example: API endpoint migrations with strict method requirements.
	ErrPermanentRedirect = errors.New("permanent redirect")

	// ErrBadRequest represents an HTTP 400 Bad Request error.
	// This error is returned when the server cannot process the request due to malformed syntax,
	// invalid parameters, or client-side errors in the request payload.
	// Use this when validating input fails or the request format is incorrect.
	// Example: malformed JSON in an API request.
	ErrBadRequest = errors.New("bad request")

	// ErrUnauthorized represents an HTTP 401 Unauthorized error.
	// This error indicates that the request lacks valid authentication credentials (e.g., token, username/password).
	// Return this when a user attempts to access a protected resource without proper authorization.
	// Note: Often paired with a WWW-Authenticate header in HTTP responses.
	// Example: missing or invalid API key.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden represents an HTTP 403 Forbidden error.
	// This error signifies that the server understood the request, but the client is not allowed to access
	// the resource, even with valid credentials (e.g., insufficient permissions).
	// Use this for access control violations.
	// Example: a user trying to access an admin-only endpoint.
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound represents an HTTP 404 Not Found error.
	// This error indicates that the requested resource could not be found on the server.
	// Commonly used when a webpage, file, or API endpoint does not exist or has been removed.
	// Example: requesting a non-existent user profile by ID.
	ErrNotFound = errors.New("not found")

	// ErrMethodNotAllowed represents an HTTP 405 Method Not Allowed error.
	// This error is returned when the HTTP method used (e.g., POST, GET) is not supported for the requested resource.
	// Useful in APIs to enforce allowed methods on endpoints.
	// Example: sending a POST request to a GET-only resource.
	// Tip: Responses typically include an Allow header listing permitted methods.
	ErrMethodNotAllowed = errors.New("method not allowed")

	// ErrNotImplemented represents an HTTP 501 Not Implemented error.
	// This error indicates that the server does not support the functionality required to fulfill the request.
	// Often used for unimplemented features or unsupported HTTP methods in development.
	// Example: an API endpoint planned but not yet coded.
	ErrNotImplemented = errors.New("not implemented")

	// ErrContentType indicates that the content-type of the request is not supported.
	// This error is returned when the server cannot process the request due to an unsupported media type
	// in the Content-Type header (e.g., expecting JSON but receiving XML).
	// Use this in APIs or handlers requiring specific content types.
	// Example: rejecting a request with "text/plain" when "application/json" is required.
	ErrContentType = errors.New("content-type not supported")

	// ErrCors indicates that a cross-origin request was blocked.
	// This error is triggered when a request violates Cross-Origin Resource Sharing (CORS) policies,
	// such as mismatched origins or missing CORS headers.
	// Common in web applications interacting with APIs from different domains.
	// Example: a frontend app on localhost trying to access a restricted API.
	ErrCors = errors.New("cross-origin request blocked")

	// ErrCallback indicates an error related to a callback operation.
	// This error is used when a callback function or mechanism fails, such as in asynchronous operations
	// or webhook processing.
	// Note: The specific cause may need additional context (consider wrapping with fmt.Errorf if needed).
	// Example: invalid callback URL provided in an API request.
	ErrCallback = errors.New("callback error")

	// ErrUnexpected indicates an unexpected error occurred.
	// This is a catch-all error for unanticipated issues that don’t fit specific categories,
	// such as internal server errors or unhandled edge cases.
	// Use sparingly; prefer specific errors when possible.
	// Example: a third-party service unexpectedly fails.
	ErrUnexpected = errors.New("unexpected error")

	// ErrNotVerified indicates that an object or entity has not been verified.
	// This error is returned when a required verification step (e.g., email, identity) has not been completed.
	// Useful in workflows requiring validation or authentication checks.
	// Example: attempting to use an unverified user account.
	ErrNotVerified = errors.New("object not verified")

	// ErrInvalid indicates that an object or input is invalid.
	// This error signifies that the provided data or resource does not meet expected criteria,
	// such as format, range, or logical constraints.
	// Use this for general validation failures not covered by specific errors like ErrBadRequest.
	// Example: an invalid date string in a form submission.
	ErrInvalid = errors.New("object invalid")
)
