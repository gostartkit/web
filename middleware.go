package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

// DefaultRequestIDHeader is the default request and response header used by RequestID middleware.
//
// Pass an empty header name to RequestID to use this value.
const DefaultRequestIDHeader = "X-Request-Id"

type requestIDContextKey struct{}

var _requestIDSeq atomic.Uint64

// RecoverOptions configures panic recovery middleware behavior.
//
// Zero values are valid: DefaultStatus falls back to 500 Internal Server Error
// and DefaultBody falls back to the standard HTTP status text.
type RecoverOptions struct {
	// Handler receives the active request context and the recovered panic value.
	// Return nil when the handler has fully handled the response, or return an
	// error to let the framework error writer produce the response.
	Handler func(c *Ctx, recovered any) error

	// DefaultStatus is used when Handler is nil.
	DefaultStatus int

	// DefaultBody is written by the default panic response when Handler is nil.
	DefaultBody string
}

// AccessLogEntry is the payload emitted by AccessLogWithOptions after a request.
//
// Status is inferred from the returned handler value/error unless a custom
// StatusMapper is configured.
type AccessLogEntry struct {
	// Status is the HTTP status associated with the request result.
	Status int

	// Duration is the elapsed time measured around the wrapped handler.
	Duration time.Duration

	// Error is the error returned by the wrapped handler, if any.
	Error error
}

// AccessLogOptions configures access logging middleware behavior.
//
// Use AccessLog for the common callback shape, or AccessLogWithOptions when
// tests need a deterministic clock or applications need custom status mapping.
type AccessLogOptions struct {
	// Log is called once after downstream request handling completes.
	Log func(c *Ctx, entry AccessLogEntry)

	// Now returns the current time. When nil, time.Now is used.
	Now func() time.Time

	// StatusMapper maps the handler result to a status code for logging.
	// When nil, framework defaults are used.
	StatusMapper func(c *Ctx, val any, err error) int
}

// RequestIDFromContext returns the request ID stored by RequestID middleware.
//
// It is useful from lower-level code that only receives context.Context. The
// function returns an empty string when ctx is nil or no request ID is present.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(requestIDContextKey{}).(string)
	return id
}

// RequestID injects a request ID into the request context and response headers.
//
// If the incoming request already contains the header, it is preserved. When
// nextID is nil, a compact monotonic base-36 ID is generated. Handlers can read
// the value with Ctx.RequestID or RequestIDFromContext.
func RequestID(header string, nextID func() string) Middleware {
	if header == "" {
		header = DefaultRequestIDHeader
	}
	if nextID == nil {
		nextID = func() string {
			return strconv.FormatUint(_requestIDSeq.Add(1), 36)
		}
	}

	return func(next Next) Next {
		return func(c *Ctx) (any, error) {
			id := c.GetHeader(header)
			if id == "" {
				id = nextID()
			}
			if id != "" {
				c.SetHeader(header, id)
				c.r = c.r.WithContext(context.WithValue(c.r.Context(), requestIDContextKey{}, id))
			}
			return next(c)
		}
	}
}

// Recover converts panics in downstream middleware and handlers into framework errors.
//
// If handler is nil, the middleware produces a default 500 response. If handler
// returns nil, the panic is considered handled and no additional body is written.
// Install this middleware with app.Use or group.Use before registering routes.
func Recover(handler func(c *Ctx, recovered any) error) Middleware {
	return RecoverWithOptions(RecoverOptions{
		Handler:       handler,
		DefaultStatus: http.StatusInternalServerError,
		DefaultBody:   "INTERNALSERVERERROR",
	})
}

// RecoverWithOptions converts panics in downstream middleware and handlers using opts.
//
// This variant is useful when the default panic status/body should be customized
// or when a custom Handler needs access to the recovered value.
func RecoverWithOptions(opts RecoverOptions) Middleware {
	status := opts.DefaultStatus
	if status == 0 {
		status = http.StatusInternalServerError
	}
	body := opts.DefaultBody
	if body == "" {
		body = http.StatusText(status)
	}

	return func(next Next) Next {
		return func(c *Ctx) (val any, err error) {
			defer func() {
				if recovered := recover(); recovered != nil {
					if opts.Handler != nil {
						err = opts.Handler(c, recovered)
						if err == nil {
							err = handledError(status)
						}
						return
					}

					err = NewErrFn(status, body, func(w http.ResponseWriter, r *http.Request) error {
						writeCodeByMedia(w, c.responseMediaType(), status)
						return c.write(body)
					})
				}
			}()

			return next(c)
		}
	}
}

func handledError(code int) error {
	return NewErrFn(code, http.StatusText(code), func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})
}

// Timeout applies a cooperative deadline to request processing.
//
// The middleware replaces the request context with one that has the configured
// deadline. Handlers and downstream services must observe c.Context() for work to
// stop promptly. If the deadline expires, ErrRequestTimeout is returned.
func Timeout(d time.Duration) Middleware {
	if d <= 0 {
		return func(next Next) Next { return next }
	}

	return func(next Next) Next {
		return func(c *Ctx) (any, error) {
			ctx, cancel := context.WithTimeout(c.r.Context(), d)
			defer cancel()

			original := c.r
			c.r = c.r.WithContext(ctx)
			defer func() { c.r = original }()

			val, err := next(c)
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, ErrRequestTimeout
			}
			return val, err
		}
	}
}

// AccessLog calls fn after request handling with an inferred status, duration, and error.
//
// It is intentionally lightweight: if fn is nil, the middleware is a no-op.
// For example, install it before routes to emit one structured log event per
// request.
func AccessLog(fn func(c *Ctx, status int, d time.Duration, err error)) Middleware {
	if fn == nil {
		return func(next Next) Next { return next }
	}

	return AccessLogWithOptions(AccessLogOptions{
		Log: func(c *Ctx, entry AccessLogEntry) {
			fn(c, entry.Status, entry.Duration, entry.Error)
		},
	})
}

// AccessLogWithOptions calls opts.Log after request handling with configurable timing/status logic.
//
// Use StatusMapper to align logs with custom response semantics, or Now to make
// timing deterministic in tests. If opts.Log is nil, the middleware is a no-op.
func AccessLogWithOptions(opts AccessLogOptions) Middleware {
	if opts.Log == nil {
		return func(next Next) Next { return next }
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	statusMapper := opts.StatusMapper
	if statusMapper == nil {
		statusMapper = func(c *Ctx, val any, err error) int {
			return statusFromResult(c, val, err)
		}
	}

	return func(next Next) Next {
		return func(c *Ctx) (any, error) {
			start := now()
			val, err := next(c)
			opts.Log(c, AccessLogEntry{
				Status:   statusMapper(c, val, err),
				Duration: now().Sub(start),
				Error:    err,
			})
			return val, err
		}
	}
}

func statusFromResult(c *Ctx, val any, err error) int {
	if err != nil {
		return errCode(err)
	}
	if c != nil && c.statusCode != 0 {
		return c.statusCode
	}
	if val == nil {
		return http.StatusNoContent
	}
	return http.StatusOK
}
