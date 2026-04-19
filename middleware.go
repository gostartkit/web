package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

// DefaultRequestIDHeader is the default request/response header used by RequestID middleware.
const DefaultRequestIDHeader = "X-Request-Id"

type requestIDContextKey struct{}

var _requestIDSeq atomic.Uint64

// RecoverOptions configures panic recovery middleware behavior.
type RecoverOptions struct {
	Handler       func(c *Ctx, recovered any) error
	DefaultStatus int
	DefaultBody   string
}

// AccessLogEntry is the payload emitted by AccessLogWithOptions.
type AccessLogEntry struct {
	Status   int
	Duration time.Duration
	Error    error
}

// AccessLogOptions configures access logging middleware behavior.
type AccessLogOptions struct {
	Log          func(c *Ctx, entry AccessLogEntry)
	Now          func() time.Time
	StatusMapper func(c *Ctx, val any, err error) int
}

// RequestIDFromContext returns the request ID stored by RequestID middleware.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(requestIDContextKey{}).(string)
	return id
}

// RequestID injects a request ID into the request context and response headers.
// If the incoming request already contains the header, it is preserved.
// When nextID is nil, a compact monotonic ID is generated.
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

// Recover converts panics in downstream middleware/handlers into framework errors.
// If handler is nil, a default 500 JSON response is produced.
// If handler returns nil, the panic is assumed to have been fully handled.
func Recover(handler func(c *Ctx, recovered any) error) Middleware {
	return RecoverWithOptions(RecoverOptions{
		Handler:       handler,
		DefaultStatus: http.StatusInternalServerError,
		DefaultBody:   "INTERNALSERVERERROR",
	})
}

// RecoverWithOptions converts panics in downstream middleware/handlers into framework errors.
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
// Handlers that observe c.Context() can stop work early and return promptly.
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

// AccessLog calls fn after request handling with an inferred status code and duration.
// The inferred status matches the framework's default success/error semantics.
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

// AccessLogWithOptions calls opts.Log after request handling with an inferred status code and duration.
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
	if c != nil && c.statusSet {
		return c.statusCode
	}
	if val == nil {
		return http.StatusNoContent
	}
	return http.StatusOK
}
