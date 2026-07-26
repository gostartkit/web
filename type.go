package web

import (
	"net/http"
	"sync"
)

// IRelease is implemented by handler return values that need cleanup after the
// response body has been written.
//
// ServeHTTP calls Release after the framework has finished with the value,
// including error paths. This is useful for pooled DTOs or buffers that should
// be returned to an object pool after encoding.
type IRelease interface {
	Release()
}

// Next is the framework handler signature.
//
// The returned value is encoded according to the request Accept header. Returning
// nil with a nil error produces 204 No Content unless the handler already wrote
// the response manually. Returning a non-nil error delegates to the configured
// ErrorHandler or the framework default error writer.
//
// Example:
//
//	app.Get("/health", func(c *web.Ctx) (any, error) {
//		return map[string]string{"status": "ok"}, nil
//	})
type Next func(c *Ctx) (any, error)

// Option configures an Application at construction time.
//
// Options are applied by New before routes are registered. They provide a
// declarative alternative to calling setters immediately after construction.
type Option func(*Application)

// Cors is the hook used by Application to populate CORS headers for automatic
// OPTIONS responses.
//
// The set function writes response headers, origin is the incoming Origin header,
// and allow contains the methods allowed for the requested path.
type Cors func(set func(key string, value string), origin string, allow []string)

// Fn is a low-level HTTP callback used by redirect and custom error responses.
//
// It receives the original ResponseWriter and Request so it can interoperate
// directly with net/http helpers such as http.Redirect.
type Fn func(w http.ResponseWriter, r *http.Request) error

// Panic is called when Application recovers a panic outside route middleware.
//
// Use Recover or RecoverWithOptions middleware when you want route-local panic
// handling that participates in framework error semantics.
type Panic func(http.ResponseWriter, *http.Request, any)

// ErrorHandler handles errors returned from route handlers.
//
// Returning nil means the handler wrote the complete error response. Returning a
// non-nil error delegates that error to the framework default writer. Install a
// handler with SetErrorHandler or WithErrorHandler.
type ErrorHandler func(c *Ctx, err error) error

// Middleware wraps a Next handler.
//
// Middleware is applied at route registration time, so routes do not pay a
// request-time composition cost. Middleware should call next(c) unless it fully
// handles the request.
type Middleware func(Next) Next

// Chain is an ordered list of Middleware values.
//
// Chains are mostly used internally when combining application-level, group-level,
// and route-level middleware.
type Chain []Middleware

// RouteGroup groups routes under a shared path prefix and middleware chain.
//
// Groups are lightweight registration helpers. Nested groups combine their
// prefixes and middleware when routes are added.
type RouteGroup struct {
	mu         sync.Mutex
	app        *Application
	prefix     string
	middleware Chain
}

// Reader decodes a request body into v for a registered media type.
//
// Register custom readers with Application.RegisterReader. When no custom reader
// is registered, TryParseBody uses the built-in JSON, GOB, and XML decoders.
type Reader func(c *Ctx, v any) error

// Writer encodes v into the response for a registered media type.
//
// Register custom writers with Application.RegisterWriter. When no custom writer
// is registered, framework-managed responses use the built-in JSON, GOB, XML,
// binary, or Avro writers.
type Writer func(c *Ctx, v any) error

// RawBody is an explicit opt-in container for raw HTTP response bytes.
//
// Passing *RawBody to DoReqWithClient or related helpers bypasses JSON decoding
// and copies the response body bytes into the slice.
type RawBody []byte

// AvroMarshaler allows custom zero-reflection avro serialization in hot paths.
//
// Values implementing AvroMarshaler can be returned from handlers when the
// request Accept header selects application/x-avro.
type AvroMarshaler interface {
	MarshalAvro() ([]byte, error)
}

// Param is a single route parameter key/value pair.
//
// Params are normally accessed through Ctx.Param and typed helpers such as
// Ctx.ParamUint64 rather than handled directly.
type Param struct {
	Key   string
	Value string
}

// Params is a list of route parameters captured during routing.
//
// The framework pools Params internally on the request path. If you use Params
// directly, treat it as request-scoped data.
type Params []Param

// Val returns the value for name from the parameter list.
//
// When duplicate names exist, the last matching parameter wins. An empty string
// is returned when the parameter is not present.
func (o *Params) Val(name string) string {
	if o == nil {
		return ""
	}
	for i := len(*o) - 1; i >= 0; i-- {
		if (*o)[i].Key == name {
			return (*o)[i].Value
		}
	}
	return ""
}
