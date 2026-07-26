package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
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

// CORSOptions configures CORS headers for both automatic OPTIONS responses and route middleware.
//
// Use NewCORS with Application.SetCORS or WithCORS for the framework's automatic
// OPTIONS handling. Use CORSMiddleware when matched routes should also emit CORS
// headers on normal responses.
type CORSOptions struct {
	// AllowOrigins lists the allowed origins. A literal "*" allows any origin.
	// When AllowCredentials is true and "*" is configured, the request Origin is
	// echoed back instead because wildcard credentials are not valid CORS.
	AllowOrigins []string

	// AllowOriginFunc optionally accepts or rejects an origin dynamically.
	// It is checked after AllowOrigins.
	AllowOriginFunc func(origin string) bool

	// AllowMethods overrides the Access-Control-Allow-Methods header. When empty,
	// NewCORS uses the framework's known route methods and CORSMiddleware echoes
	// the requested preflight method when available.
	AllowMethods []string

	// AllowHeaders sets Access-Control-Allow-Headers for preflight responses.
	// When empty, CORSMiddleware echoes Access-Control-Request-Headers.
	AllowHeaders []string

	// ExposeHeaders sets Access-Control-Expose-Headers on non-preflight responses.
	ExposeHeaders []string

	// AllowCredentials sets Access-Control-Allow-Credentials to true.
	AllowCredentials bool

	// MaxAge controls the Access-Control-Max-Age header on preflight responses.
	MaxAge time.Duration

	// PassthroughOptions controls whether CORSMiddleware continues into the next
	// handler for preflight requests. By default preflight requests are completed
	// by the middleware with 204 No Content.
	PassthroughOptions bool
}

// SecurityHeadersOptions configures SecurityHeadersWithOptions.
//
// When UseDefaults is true, unset fields fall back to the middleware defaults:
// X-Content-Type-Options=nosniff, X-Frame-Options=DENY, and
// Referrer-Policy=no-referrer.
type SecurityHeadersOptions struct {
	// UseDefaults applies the standard default header set when individual values
	// are left empty.
	UseDefaults bool

	// ContentTypeOptions sets X-Content-Type-Options.
	ContentTypeOptions string

	// FrameOptions sets X-Frame-Options.
	FrameOptions string

	// ReferrerPolicy sets Referrer-Policy.
	ReferrerPolicy string

	// ContentSecurityPolicy sets Content-Security-Policy.
	ContentSecurityPolicy string

	// StrictTransportSecurity sets Strict-Transport-Security.
	StrictTransportSecurity string

	// CrossOriginOpenerPolicy sets Cross-Origin-Opener-Policy.
	CrossOriginOpenerPolicy string

	// CrossOriginResourcePolicy sets Cross-Origin-Resource-Policy.
	CrossOriginResourcePolicy string

	// PermissionsPolicy sets Permissions-Policy.
	PermissionsPolicy string
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

// NewCORS returns an Application CORS hook for automatic OPTIONS responses.
//
// Install it with WithCORS or SetCORS when the framework's built-in automatic
// OPTIONS handling should emit standard CORS headers. For matched GET/POST/etc.
// responses, combine this with CORSMiddleware when those routes should also emit
// the same origin/credentials/expose headers.
func NewCORS(opts CORSOptions) Cors {
	opts = cloneCORSOptions(opts)
	return func(set func(key string, value string), origin string, allow []string) {
		if origin == "" {
			return
		}

		header := http.Header{}
		allowedOrigin, varyOrigin, ok := resolveCORSOrigin(origin, opts)
		if !ok {
			return
		}

		if varyOrigin {
			addVaryHeader(header, "Origin")
		}
		header.Set("Access-Control-Allow-Origin", allowedOrigin)
		if opts.AllowCredentials {
			header.Set("Access-Control-Allow-Credentials", "true")
		}

		methods := opts.AllowMethods
		if len(methods) == 0 {
			methods = allow
		}
		if len(methods) > 0 {
			header.Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
		}
		if len(opts.AllowHeaders) > 0 {
			header.Set("Access-Control-Allow-Headers", strings.Join(opts.AllowHeaders, ", "))
		}
		if opts.MaxAge > 0 {
			header.Set("Access-Control-Max-Age", strconv.FormatInt(int64(opts.MaxAge/time.Second), 10))
		}

		for key, values := range header {
			for _, value := range values {
				set(key, value)
			}
		}
	}
}

// CORSMiddleware emits CORS headers for matched route responses.
//
// This middleware handles normal requests and explicit preflight routes. For the
// framework's automatic OPTIONS path, install NewCORS through SetCORS or
// WithCORS because automatic OPTIONS responses bypass route middleware.
func CORSMiddleware(opts CORSOptions) Middleware {
	opts = cloneCORSOptions(opts)
	return func(next Next) Next {
		return func(c *Ctx) (any, error) {
			origin := c.Origin()
			if origin == "" {
				return next(c)
			}

			if !applyCORSResponseHeaders(c.ResponseWriter().Header(), c.Request(), opts) {
				return next(c)
			}

			if c.Method() == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" && !opts.PassthroughOptions {
				return nil, nil
			}

			return next(c)
		}
	}
}

func cloneCORSOptions(opts CORSOptions) CORSOptions {
	opts.AllowOrigins = append([]string(nil), opts.AllowOrigins...)
	opts.AllowMethods = append([]string(nil), opts.AllowMethods...)
	opts.AllowHeaders = append([]string(nil), opts.AllowHeaders...)
	opts.ExposeHeaders = append([]string(nil), opts.ExposeHeaders...)
	return opts
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
						c.commitMedia(c.responseMediaType(), status)
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

// MaxBodyBytes limits the size of request bodies read by downstream handlers.
//
// It wraps the request body with http.MaxBytesReader. A limit <= 0 disables the
// middleware. When the body exceeds the limit, downstream readers return a
// standard MaxBytes error that the framework maps to a 400 response unless a
// custom error handler overrides it.
func MaxBodyBytes(limit int64) Middleware {
	if limit <= 0 {
		return func(next Next) Next { return next }
	}

	return func(next Next) Next {
		return func(c *Ctx) (any, error) {
			if c.r.Body != nil {
				c.r.Body = http.MaxBytesReader(c.w, c.r.Body, limit)
			}
			return next(c)
		}
	}
}

// SecurityHeaders applies a small set of common security response headers.
//
// The default headers are:
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY
//   - Referrer-Policy: no-referrer
//
// Use SecurityHeadersWithOptions when additional or custom policies are needed.
func SecurityHeaders() Middleware {
	return SecurityHeadersWithOptions(SecurityHeadersOptions{UseDefaults: true})
}

// SecurityHeadersWithOptions applies configurable security response headers.
//
// It is intentionally lightweight and only sets headers configured by opts plus
// the defaults when opts.UseDefaults is true.
func SecurityHeadersWithOptions(opts SecurityHeadersOptions) Middleware {
	if opts.UseDefaults {
		if opts.ContentTypeOptions == "" {
			opts.ContentTypeOptions = "nosniff"
		}
		if opts.FrameOptions == "" {
			opts.FrameOptions = "DENY"
		}
		if opts.ReferrerPolicy == "" {
			opts.ReferrerPolicy = "no-referrer"
		}
	}

	return func(next Next) Next {
		return func(c *Ctx) (any, error) {
			header := c.ResponseWriter().Header()
			if opts.ContentTypeOptions != "" {
				header.Set("X-Content-Type-Options", opts.ContentTypeOptions)
			}
			if opts.FrameOptions != "" {
				header.Set("X-Frame-Options", opts.FrameOptions)
			}
			if opts.ReferrerPolicy != "" {
				header.Set("Referrer-Policy", opts.ReferrerPolicy)
			}
			if opts.ContentSecurityPolicy != "" {
				header.Set("Content-Security-Policy", opts.ContentSecurityPolicy)
			}
			if opts.StrictTransportSecurity != "" {
				header.Set("Strict-Transport-Security", opts.StrictTransportSecurity)
			}
			if opts.CrossOriginOpenerPolicy != "" {
				header.Set("Cross-Origin-Opener-Policy", opts.CrossOriginOpenerPolicy)
			}
			if opts.CrossOriginResourcePolicy != "" {
				header.Set("Cross-Origin-Resource-Policy", opts.CrossOriginResourcePolicy)
			}
			if opts.PermissionsPolicy != "" {
				header.Set("Permissions-Policy", opts.PermissionsPolicy)
			}
			return next(c)
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

func resolveCORSOrigin(origin string, opts CORSOptions) (allowed string, varyOrigin bool, ok bool) {
	if origin == "" {
		return "", false, false
	}

	allowAny := len(opts.AllowOrigins) == 0 && opts.AllowOriginFunc == nil
	for _, candidate := range opts.AllowOrigins {
		if candidate == "*" {
			allowAny = true
			break
		}
		if candidate == origin {
			return origin, true, true
		}
	}

	if opts.AllowOriginFunc != nil && opts.AllowOriginFunc(origin) {
		return origin, true, true
	}

	if !allowAny {
		return "", false, false
	}
	if opts.AllowCredentials {
		return origin, true, true
	}
	return "*", false, true
}

func applyCORSResponseHeaders(header http.Header, req *http.Request, opts CORSOptions) bool {
	allowedOrigin, varyOrigin, ok := resolveCORSOrigin(req.Header.Get("Origin"), opts)
	if !ok {
		return false
	}

	if varyOrigin {
		addVaryHeader(header, "Origin")
	}
	header.Set("Access-Control-Allow-Origin", allowedOrigin)
	if opts.AllowCredentials {
		header.Set("Access-Control-Allow-Credentials", "true")
	}
	if len(opts.ExposeHeaders) > 0 && req.Method != http.MethodOptions {
		header.Set("Access-Control-Expose-Headers", strings.Join(opts.ExposeHeaders, ", "))
	}

	if req.Method == http.MethodOptions && req.Header.Get("Access-Control-Request-Method") != "" {
		if len(opts.AllowMethods) > 0 {
			header.Set("Access-Control-Allow-Methods", strings.Join(opts.AllowMethods, ", "))
		} else if method := req.Header.Get("Access-Control-Request-Method"); method != "" {
			addVaryHeader(header, "Access-Control-Request-Method")
			header.Set("Access-Control-Allow-Methods", method)
		}

		if len(opts.AllowHeaders) > 0 {
			header.Set("Access-Control-Allow-Headers", strings.Join(opts.AllowHeaders, ", "))
		} else if headers := req.Header.Get("Access-Control-Request-Headers"); headers != "" {
			addVaryHeader(header, "Access-Control-Request-Headers")
			header.Set("Access-Control-Allow-Headers", headers)
		}

		if opts.MaxAge > 0 {
			header.Set("Access-Control-Max-Age", strconv.FormatInt(int64(opts.MaxAge/time.Second), 10))
		}
	}

	return true
}

func addVaryHeader(header http.Header, value string) {
	existing := header.Values("Vary")
	for _, raw := range existing {
		for _, item := range strings.Split(raw, ",") {
			if strings.TrimSpace(item) == value {
				return
			}
		}
	}
	header.Add("Vary", value)
}
