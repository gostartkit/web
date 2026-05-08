package web

import (
	"bufio"
	"bytes"
	"context"
	"encoding"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var (
	_ctxPool = sync.Pool{
		New: func() any {
			c := &Ctx{}
			return c
		}}
	_bodyReadBufferPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
	_copyBufPool = sync.Pool{
		New: func() any {
			b := make([]byte, 32*1024)
			return &b
		},
	}
)

// createCtx returns a new instance of web.Ctx, initialized with the given HTTP response writer, request, and parameters.
func createCtx(app *Application, w http.ResponseWriter, r *http.Request, params *Params) *Ctx {
	c := _ctxPool.Get().(*Ctx)
	c.app = app
	c.w = w
	c.r = r
	c.param = params
	return c
}

func createCtxWithRouteMatch(app *Application, w http.ResponseWriter, r *http.Request, match *routeMatch) *Ctx {
	c := _ctxPool.Get().(*Ctx)
	c.app = app
	c.w = w
	c.r = r
	c.routeParamNames = match.paramNames
	c.routeParamCount = match.paramCount
	c.routeParamValue0 = match.paramValue0
	c.routeParamValue1 = match.paramValue1
	c.routeParamValue2 = match.paramValue2
	c.routeParamExtraValues = match.paramExtraValues
	return c
}

// releaseCtx puts the context object back into the pool for reuse.
func releaseCtx(c *Ctx) {
	if c != nil {
		*c = Ctx{}
		_ctxPool.Put(c)
	}
}

// Ctx represents one HTTP request/response exchange.
//
// A Ctx is pooled and reused by the framework. Do not store it beyond the
// lifetime of the handler. Use Request, ResponseWriter, and Context when you
// need direct access to standard net/http primitives.
type Ctx struct {
	app                    *Application
	w                      http.ResponseWriter
	r                      *http.Request
	param                  *Params
	routeParamNames        []string
	routeParamExtraValues  *[]string
	query                  url.Values
	userId                 uint64
	formDataState          uint8
	routeParamCount        uint16
	statusCode             int
	responseCommitted      bool
	acceptType             mediaType
	acceptTypeCached       bool
	contentTypeValue       string
	contentTypeValueCached bool
	contentType            mediaType
	contentTypeCached      bool
	routeParamValue0       string
	routeParamValue1       string
	routeParamValue2       string
}

// Init stores the authenticated user ID on the request context.
//
// The framework does not call Init automatically; applications or middleware can
// call it after authentication so handlers can read the value with UserId.
func (c *Ctx) Init(userId uint64) {
	c.userId = userId
}

// Request returns the underlying *http.Request.
//
// Use this when a handler needs access to standard library request fields or
// APIs that are not wrapped by Ctx.
func (c *Ctx) Request() *http.Request {
	return c.r
}

// ResponseWriter returns the underlying http.ResponseWriter.
//
// Prefer Ctx helpers for common writes, but use ResponseWriter when integrating
// lower-level net/http code.
func (c *Ctx) ResponseWriter() http.ResponseWriter {
	return c.w
}

// Header implements http.ResponseWriter and returns the response header map.
//
// Mutate this map before the response is committed. After WriteHeader or Write,
// behavior follows the underlying ResponseWriter.
func (c *Ctx) Header() http.Header {
	return c.w.Header()
}

// Write implements http.ResponseWriter and writes bytes to the response body.
//
// If no status was committed, Write commits status 200 OK unless SetStatus was
// called first. Returning a non-nil value from the handler after calling Write is
// not recommended because the response has already been handled manually.
func (c *Ctx) Write(p []byte) (int, error) {
	if !c.responseCommitted {
		code := c.statusCode
		if code == 0 {
			code = http.StatusOK
			c.statusCode = code
		}
		c.responseCommitted = true
		if code != http.StatusOK {
			c.w.WriteHeader(code)
		}
	}
	return c.w.Write(p)
}

// WriteHeader implements http.ResponseWriter and commits the response status.
//
// Additional calls after the response is committed are ignored by Ctx.
func (c *Ctx) WriteHeader(statusCode int) {
	if c.responseCommitted {
		return
	}
	c.statusCode = statusCode
	c.responseCommitted = true
	c.w.WriteHeader(statusCode)
}

// SetStatus sets the response status code for framework-managed writes without
// immediately committing the response.
//
// Use SetStatus when returning a value and letting the framework encode it:
//
//	c.SetStatus(http.StatusCreated)
//	return user, nil
func (c *Ctx) SetStatus(statusCode int) {
	c.statusCode = statusCode
}

// QueryValues returns the parsed query string values for the request URL.
//
// Values are parsed lazily and cached on the Ctx for the remainder of the
// request.
func (c *Ctx) QueryValues() url.Values {
	if c.query == nil {
		c.query = c.r.URL.Query()
	}
	return c.query
}

// UserId returns the user ID previously stored with Init.
//
// A zero value means no user ID was set, unless zero is a valid user ID in the
// application domain.
func (c *Ctx) UserId() uint64 {
	return c.userId
}

// RequestID returns the request ID injected by RequestID middleware.
//
// It returns an empty string when the middleware is not installed or no request
// is attached to the Ctx.
func (c *Ctx) RequestID() string {
	if c == nil || c.r == nil {
		return ""
	}
	return RequestIDFromContext(c.r.Context())
}

// Param returns a route parameter captured by name.
//
// Parameters come from :name and *name route segments. If the parameter is not
// present, Param returns an empty string. For numeric parameters, prefer typed
// helpers such as ParamUint64 to avoid repeated manual parsing.
func (c *Ctx) Param(name string) string {
	if c.routeParamCount != 0 {
		names := c.routeParamNames
		switch c.routeParamCount {
		case 1:
			if names[0] == name {
				return c.routeParamValue0
			}
			return ""
		case 2:
			if names[1] == name {
				return c.routeParamValue1
			}
			if names[0] == name {
				return c.routeParamValue0
			}
			return ""
		case 3:
			if names[2] == name {
				return c.routeParamValue2
			}
			if names[1] == name {
				return c.routeParamValue1
			}
			if names[0] == name {
				return c.routeParamValue0
			}
			return ""
		}
		for i := len(names) - 1; i >= 0; i-- {
			if names[i] == name {
				switch i {
				case 0:
					return c.routeParamValue0
				case 1:
					return c.routeParamValue1
				case 2:
					return c.routeParamValue2
				default:
					return (*c.routeParamExtraValues)[i-3]
				}
			}
		}
		return ""
	}
	if c.param == nil {
		return ""
	}
	return c.param.Val(name)
}

// Query returns the first query string value for name.
//
// It follows url.Values.Get semantics: an absent key returns an empty string.
// Use QueryValues when you need all values for a repeated query key.
func (c *Ctx) Query(name string) string {
	return c.QueryValues().Get(name)
}

// Form returns the first form value for name.
//
// It delegates to http.Request.FormValue, so it may parse URL-encoded or
// multipart form data. Query parameters are considered by the standard library
// when FormValue is used.
func (c *Ctx) Form(name string) string {
	return c.r.FormValue(name)
}

// PostForm returns the first POST form value for name.
//
// It delegates to http.Request.PostFormValue and does not fall back to URL query
// parameters.
func (c *Ctx) PostForm(name string) string {
	return c.r.PostFormValue(name)
}

// FormFile returns the first uploaded file for key.
//
// It delegates to http.Request.FormFile, which parses multipart form data when
// needed. The caller is responsible for closing the returned file.
func (c *Ctx) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.r.FormFile(key)
}

// Host returns the request Host value.
//
// This is the host from http.Request.Host, not necessarily a Host header looked
// up through Header.Get.
func (c *Ctx) Host() string {
	return c.r.Host
}

// Path returns the request URL path.
//
// The value is the routed path used by the framework, such as /users/42.
func (c *Ctx) Path() string {
	return c.r.URL.Path
}

// Body returns the request body stream.
//
// Reading from Body consumes it. Use TryParseBody or TryParseJSONBodyFast for
// common structured request decoding.
func (c *Ctx) Body() io.ReadCloser {
	return c.r.Body
}

// Method returns the HTTP method used for the request.
//
// Examples include GET, POST, PUT, PATCH, DELETE, and custom methods registered
// through Application.Handle.
func (c *Ctx) Method() string {
	return c.r.Method
}

// RemoteAddr returns the remote network address reported by net/http.
//
// The value may include a port. In deployments behind proxies, use trusted proxy
// headers at the application layer if you need the original client IP.
func (c *Ctx) RemoteAddr() string {
	return c.r.RemoteAddr
}

// BearerToken returns the bearer token from the Authorization header.
//
// It returns an empty string unless the header is in the form
// "Bearer <token>".
func (c *Ctx) BearerToken() string {
	return bearerToken(c.GetHeader("Authorization"))
}

// Origin returns the request Origin header.
func (c *Ctx) Origin() string {
	return c.GetHeader("Origin")
}

// SetOrigin sets the Access-Control-Allow-Origin response header.
func (c *Ctx) SetOrigin(origin string) {
	c.SetHeader("Access-Control-Allow-Origin", origin)
}

// AllowCredentials sets Access-Control-Allow-Credentials to true.
func (c *Ctx) AllowCredentials() {
	c.SetHeader("Access-Control-Allow-Credentials", "true")
}

// UserAgent returns the request User-Agent header.
func (c *Ctx) UserAgent() string {
	return c.GetHeader("User-Agent")
}

// IsAjax reports whether X-Requested-With is XMLHttpRequest.
//
// This is a legacy browser convention and is not a general-purpose request type
// detector.
func (c *Ctx) IsAjax() bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

// IsFormData reports whether the request Content-Type is a supported form type.
//
// Deprecated: use IsForm instead.
func (c *Ctx) IsFormData() bool {
	return c.IsForm()
}

// IsForm reports whether the request Content-Type is a form submission.
//
// It returns true for application/x-www-form-urlencoded and multipart/form-data.
// The result is cached on the Ctx after the first call.
func (c *Ctx) IsForm() bool {

	if c.formDataState > 0 {
		return c.formDataState == 1
	}

	ct := c.ContentType()

	isForm := ct != "" && (strings.HasPrefix(ct, "application/x-www-form-urlencoded") || strings.HasPrefix(ct, "multipart/form-data"))

	if isForm {
		c.formDataState = 1
	} else {
		c.formDataState = 255
	}

	return isForm
}

// TryParseBody decodes the request body into val based on Content-Type.
//
// JSON decoding rejects unknown fields. GOB and XML use the standard library
// decoders. Custom readers registered with RegisterReader take precedence for
// their media type. ErrContentType is returned for unsupported media types.
func (c *Ctx) TryParseBody(val any) error {

	if c.r == nil || c.r.Body == nil {
		return io.EOF
	}

	switch c.requestMediaType() {
	case mediaJSON:
		if c.app != nil && c.app.hasReaders {
			if reader := c.app.readers[mediaJSON]; reader != nil {
				return reader(c, val)
			}
		}
		dec := json.NewDecoder(c.r.Body)
		dec.DisallowUnknownFields()
		return dec.Decode(val)
	case mediaGOB:
		if c.app != nil && c.app.hasReaders {
			if reader := c.app.readers[mediaGOB]; reader != nil {
				return reader(c, val)
			}
		}
		dec := gob.NewDecoder(c.r.Body)
		return dec.Decode(val)
	case mediaOctetStream:
		if c.app != nil && c.app.hasReaders {
			if reader := c.app.readers[mediaOctetStream]; reader != nil {
				return reader(c, val)
			}
		}
		return ErrContentType
	case mediaXML:
		if c.app != nil && c.app.hasReaders {
			if reader := c.app.readers[mediaXML]; reader != nil {
				return reader(c, val)
			}
		}
		dec := xml.NewDecoder(c.r.Body)
		return dec.Decode(val)
	default:
		return ErrContentType
	}
}

// TryParseJSONBodyFast parses a JSON request body using a pooled buffer and
// json.Unmarshal. This is faster than TryParseBody for common JSON payloads,
// but unlike TryParseBody it does not reject unknown fields.
//
// Use this in hot paths where standard json.Unmarshal semantics are acceptable.
func (c *Ctx) TryParseJSONBodyFast(val any) error {
	if c.r == nil || c.r.Body == nil {
		return io.EOF
	}

	buf := _bodyReadBufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	_, err := buf.ReadFrom(c.r.Body)
	if err == nil {
		err = json.Unmarshal(buf.Bytes(), val)
	}

	buf.Reset()
	_bodyReadBufferPool.Put(buf)
	return err
}

// TryParseParam parses the named route parameter into val.
//
// val must be a pointer to one of the types supported by TryParse, such as
// *int, *uint64, *bool, or a supported slice pointer.
func (c *Ctx) TryParseParam(name string, val any) error {
	return TryParse(c.Param(name), val)
}

// TryParseQuery parses the named query string value into val.
//
// Missing query values are treated the same way as TryParse receiving an empty
// string.
func (c *Ctx) TryParseQuery(name string, val any) error {
	return TryParse(c.Query(name), val)
}

// TryParseForm parses the named form value into val.
//
// It uses Form, so it follows http.Request.FormValue behavior for form parsing.
func (c *Ctx) TryParseForm(name string, val any) error {
	return TryParse(c.Form(name), val)
}

// ParamInt parses the named route parameter as int.
func (c *Ctx) ParamInt(name string) (int, error) {
	return TryInt(c.Param(name))
}

// ParamUint parses the named route parameter as uint.
func (c *Ctx) ParamUint(name string) (uint, error) {
	return TryUint(c.Param(name))
}

// ParamInt8 parses the named route parameter as int8.
func (c *Ctx) ParamInt8(name string) (int8, error) {
	return TryInt8(c.Param(name))
}

// ParamUint8 parses the named route parameter as uint8.
func (c *Ctx) ParamUint8(name string) (uint8, error) {
	return TryUint8(c.Param(name))
}

// ParamInt16 parses the named route parameter as int16.
func (c *Ctx) ParamInt16(name string) (int16, error) {
	return TryInt16(c.Param(name))
}

// ParamUint16 parses the named route parameter as uint16.
func (c *Ctx) ParamUint16(name string) (uint16, error) {
	return TryUint16(c.Param(name))
}

// ParamInt32 parses the named route parameter as int32.
func (c *Ctx) ParamInt32(name string) (int32, error) {
	return TryInt32(c.Param(name))
}

// ParamUint32 parses the named route parameter as uint32.
func (c *Ctx) ParamUint32(name string) (uint32, error) {
	return TryUint32(c.Param(name))
}

// ParamInt64 parses the named route parameter as int64.
func (c *Ctx) ParamInt64(name string) (int64, error) {
	return TryInt64(c.Param(name))
}

// ParamUint64 parses the named route parameter as uint64.
//
// This is the preferred helper for numeric IDs in hot paths because it avoids
// intermediate allocations and uses the package's fast integer parser.
func (c *Ctx) ParamUint64(name string) (uint64, error) {
	return TryUint64(c.Param(name))
}

// ParamFloat32 parses the named route parameter as float32.
func (c *Ctx) ParamFloat32(name string) (float32, error) {
	return TryFloat32(c.Param(name))
}

// ParamFloat64 parses the named route parameter as float64.
func (c *Ctx) ParamFloat64(name string) (float64, error) {
	return TryFloat64(c.Param(name))
}

// ParamBool parses the named route parameter as bool.
func (c *Ctx) ParamBool(name string) (bool, error) {
	return TryBool(c.Param(name))
}

// QueryInt parses the named query value as int.
func (c *Ctx) QueryInt(name string) (int, error) {
	return TryInt(c.Query(name))
}

// QueryUint parses the named query value as uint.
func (c *Ctx) QueryUint(name string) (uint, error) {
	return TryUint(c.Query(name))
}

// QueryInt8 parses the named query value as int8.
func (c *Ctx) QueryInt8(name string) (int8, error) {
	return TryInt8(c.Query(name))
}

// QueryUint8 parses the named query value as uint8.
func (c *Ctx) QueryUint8(name string) (uint8, error) {
	return TryUint8(c.Query(name))
}

// QueryInt16 parses the named query value as int16.
func (c *Ctx) QueryInt16(name string) (int16, error) {
	return TryInt16(c.Query(name))
}

// QueryUint16 parses the named query value as uint16.
func (c *Ctx) QueryUint16(name string) (uint16, error) {
	return TryUint16(c.Query(name))
}

// QueryInt32 parses the named query value as int32.
func (c *Ctx) QueryInt32(name string) (int32, error) {
	return TryInt32(c.Query(name))
}

// QueryUint32 parses the named query value as uint32.
func (c *Ctx) QueryUint32(name string) (uint32, error) {
	return TryUint32(c.Query(name))
}

// QueryInt64 parses the named query value as int64.
func (c *Ctx) QueryInt64(name string) (int64, error) {
	return TryInt64(c.Query(name))
}

// QueryUint64 parses the named query value as uint64.
func (c *Ctx) QueryUint64(name string) (uint64, error) {
	return TryUint64(c.Query(name))
}

// QueryFloat32 parses the named query value as float32.
func (c *Ctx) QueryFloat32(name string) (float32, error) {
	return TryFloat32(c.Query(name))
}

// QueryFloat64 parses the named query value as float64.
func (c *Ctx) QueryFloat64(name string) (float64, error) {
	return TryFloat64(c.Query(name))
}

// QueryBool parses the named query value as bool.
func (c *Ctx) QueryBool(name string) (bool, error) {
	return TryBool(c.Query(name))
}

// FormInt parses the named form value as int.
func (c *Ctx) FormInt(name string) (int, error) {
	return TryInt(c.Form(name))
}

// FormUint parses the named form value as uint.
func (c *Ctx) FormUint(name string) (uint, error) {
	return TryUint(c.Form(name))
}

// FormInt8 parses the named form value as int8.
func (c *Ctx) FormInt8(name string) (int8, error) {
	return TryInt8(c.Form(name))
}

// FormUint8 parses the named form value as uint8.
func (c *Ctx) FormUint8(name string) (uint8, error) {
	return TryUint8(c.Form(name))
}

// FormInt16 parses the named form value as int16.
func (c *Ctx) FormInt16(name string) (int16, error) {
	return TryInt16(c.Form(name))
}

// FormUint16 parses the named form value as uint16.
func (c *Ctx) FormUint16(name string) (uint16, error) {
	return TryUint16(c.Form(name))
}

// FormInt32 parses the named form value as int32.
func (c *Ctx) FormInt32(name string) (int32, error) {
	return TryInt32(c.Form(name))
}

// FormUint32 parses the named form value as uint32.
func (c *Ctx) FormUint32(name string) (uint32, error) {
	return TryUint32(c.Form(name))
}

// FormInt64 parses the named form value as int64.
func (c *Ctx) FormInt64(name string) (int64, error) {
	return TryInt64(c.Form(name))
}

// FormUint64 parses the named form value as uint64.
func (c *Ctx) FormUint64(name string) (uint64, error) {
	return TryUint64(c.Form(name))
}

// FormFloat32 parses the named form value as float32.
func (c *Ctx) FormFloat32(name string) (float32, error) {
	return TryFloat32(c.Form(name))
}

// FormFloat64 parses the named form value as float64.
func (c *Ctx) FormFloat64(name string) (float64, error) {
	return TryFloat64(c.Form(name))
}

// FormBool parses the named form value as bool.
func (c *Ctx) FormBool(name string) (bool, error) {
	return TryBool(c.Form(name))
}

// PostFormInt parses the named POST form value as int.
func (c *Ctx) PostFormInt(name string) (int, error) {
	return TryInt(c.PostForm(name))
}

// PostFormUint parses the named POST form value as uint.
func (c *Ctx) PostFormUint(name string) (uint, error) {
	return TryUint(c.PostForm(name))
}

// PostFormInt8 parses the named POST form value as int8.
func (c *Ctx) PostFormInt8(name string) (int8, error) {
	return TryInt8(c.PostForm(name))
}

// PostFormUint8 parses the named POST form value as uint8.
func (c *Ctx) PostFormUint8(name string) (uint8, error) {
	return TryUint8(c.PostForm(name))
}

// PostFormInt16 parses the named POST form value as int16.
func (c *Ctx) PostFormInt16(name string) (int16, error) {
	return TryInt16(c.PostForm(name))
}

// PostFormUint16 parses the named POST form value as uint16.
func (c *Ctx) PostFormUint16(name string) (uint16, error) {
	return TryUint16(c.PostForm(name))
}

// PostFormInt32 parses the named POST form value as int32.
func (c *Ctx) PostFormInt32(name string) (int32, error) {
	return TryInt32(c.PostForm(name))
}

// PostFormUint32 parses the named POST form value as uint32.
func (c *Ctx) PostFormUint32(name string) (uint32, error) {
	return TryUint32(c.PostForm(name))
}

// PostFormInt64 parses the named POST form value as int64.
func (c *Ctx) PostFormInt64(name string) (int64, error) {
	return TryInt64(c.PostForm(name))
}

// PostFormUint64 parses the named POST form value as uint64.
func (c *Ctx) PostFormUint64(name string) (uint64, error) {
	return TryUint64(c.PostForm(name))
}

// PostFormFloat32 parses the named POST form value as float32.
func (c *Ctx) PostFormFloat32(name string) (float32, error) {
	return TryFloat32(c.PostForm(name))
}

// PostFormFloat64 parses the named POST form value as float64.
func (c *Ctx) PostFormFloat64(name string) (float64, error) {
	return TryFloat64(c.PostForm(name))
}

// PostFormBool parses the named POST form value as bool.
func (c *Ctx) PostFormBool(name string) (bool, error) {
	return TryBool(c.PostForm(name))
}

// Accept returns the request Accept header.
//
// The framework uses this value to choose the response media type for
// framework-managed writes.
func (c *Ctx) Accept() string {
	return c.GetHeader("Accept")
}

// Flusher returns the underlying http.Flusher when the writer supports it.
//
// Use this for streaming responses that need to push partial data to the client.
// It returns nil when the underlying ResponseWriter cannot flush.
func (c *Ctx) Flusher() http.Flusher {
	if flusher, ok := c.w.(http.Flusher); ok {
		return flusher
	}
	return nil
}

// Flush flushes buffered response data when the underlying writer supports it.
//
// If flushing is not supported, Flush is a no-op.
func (c *Ctx) Flush() {
	if flusher, ok := c.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijacker returns the underlying http.Hijacker when the writer supports it.
//
// This is useful for protocol upgrades or raw connection control. It returns nil
// when hijacking is unsupported.
func (c *Ctx) Hijacker() http.Hijacker {
	if hijacker, ok := c.w.(http.Hijacker); ok {
		return hijacker
	}
	return nil
}

// Hijack takes over the underlying network connection when supported.
//
// It delegates to the underlying http.Hijacker. When hijacking is unsupported,
// it returns http.ErrNotSupported.
func (c *Ctx) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := c.w.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Push initiates an HTTP/2 server push when the underlying writer supports it.
//
// It returns http.ErrNotSupported when server push is unavailable.
func (c *Ctx) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := c.w.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// Context returns the request context.
//
// Middleware such as Timeout and RequestID can replace or enrich this context
// during request handling.
func (c *Ctx) Context() context.Context {
	return c.r.Context()
}

// ContentType returns the request Content-Type header.
//
// The value is cached on first access because parsing request media type is a
// hot path for body decoding.
func (c *Ctx) ContentType() string {
	return c.requestContentType()
}

// SetContentType sets the response Content-Type header.
func (c *Ctx) SetContentType(val string) {
	c.SetHeader("Content-Type", val)
}

// SetCacheControl sets the response Cache-Control header.
func (c *Ctx) SetCacheControl(val string) {
	c.SetHeader("Cache-Control", val)
}

// SetConnection sets the response Connection header.
func (c *Ctx) SetConnection(val string) {
	c.SetHeader("Connection", val)
}

// SetVersion sets the response Version header.
func (c *Ctx) SetVersion(version string) {
	c.SetHeader("Version", version)
}

// SetCookie adds a Set-Cookie header to the response.
//
// It delegates to http.SetCookie. Invalid cookies may be silently dropped by the
// standard library.
func (c *Ctx) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.w, cookie)
}

// GetCookie returns the named cookie from the request.
//
// It delegates to http.Request.Cookie. If multiple cookies match, the standard
// library returns one of them. Missing cookies return http.ErrNoCookie.
func (c *Ctx) GetCookie(name string) (*http.Cookie, error) {
	return c.r.Cookie(name)
}

// GetHeader returns the request header value for key.
//
// It is a convenience wrapper around c.Request().Header.Get.
func (c *Ctx) GetHeader(key string) string {
	return c.r.Header.Get(key)
}

// SetHeader sets a response header.
//
// It is a convenience wrapper around c.ResponseWriter().Header().Set.
func (c *Ctx) SetHeader(key string, value string) {
	c.w.Header().Set(key, value)
}

// NoContent writes a response status without a body.
//
// A zero status defaults to 204 No Content. This helper writes immediately; if
// you prefer the framework return-value model, returning (nil, nil) already
// produces 204 when the response has not been written.
func (c *Ctx) NoContent(statusCode int) error {
	if statusCode == 0 {
		statusCode = http.StatusNoContent
	}
	c.WriteHeader(statusCode)
	return nil
}

// JSON writes a JSON response immediately.
//
// A zero status defaults to 200 OK. This bypasses Accept negotiation and always
// writes application/json. In normal framework-managed handlers, returning a
// value and optional SetStatus is the more idiomatic path.
func (c *Ctx) JSON(statusCode int, val any) error {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if !c.responseCommitted {
		writeCodeByMedia(c.w, mediaJSON, statusCode)
		c.statusCode = statusCode
		c.responseCommitted = true
	}
	return c.writeMedia(mediaJSON, val)
}

// String writes a plain-text response immediately.
//
// A zero status defaults to 200 OK. The Content-Type is set to
// text/plain; charset=utf-8 when the response has not already been committed.
func (c *Ctx) String(statusCode int, body string) error {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if !c.responseCommitted {
		c.SetContentType("text/plain; charset=utf-8")
		c.WriteHeader(statusCode)
	}
	_, err := io.WriteString(c.w, body)
	return err
}

// Blob writes raw bytes immediately with the provided content type.
//
// A zero status defaults to 200 OK. Blob is useful for pre-encoded bytes in hot
// paths where you want to avoid framework media negotiation.
func (c *Ctx) Blob(statusCode int, contentType string, body []byte) error {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if !c.responseCommitted {
		if contentType != "" {
			c.SetContentType(contentType)
		}
		c.WriteHeader(statusCode)
	}
	_, err := c.w.Write(body)
	return err
}

// write write data base on accept header
func (c *Ctx) write(val any) error {
	return c.writeMedia(c.responseMediaType(), val)
}

func (c *Ctx) writeMedia(mt mediaType, val any) error {
	switch mt {
	case mediaJSON:
		if c.app != nil && c.app.hasWriters {
			if writer := c.app.writers[mediaJSON]; writer != nil {
				return writer(c, val)
			}
		}
		return c.writeJSON(val)
	case mediaGOB:
		if c.app != nil && c.app.hasWriters {
			if writer := c.app.writers[mediaGOB]; writer != nil {
				return writer(c, val)
			}
		}
		return c.writeGOB(val)
	case mediaOctetStream:
		if c.app != nil && c.app.hasWriters {
			if writer := c.app.writers[mediaOctetStream]; writer != nil {
				return writer(c, val)
			}
		}
		return c.writeBinary(val)
	case mediaAvro:
		if c.app != nil && c.app.hasWriters {
			if writer := c.app.writers[mediaAvro]; writer != nil {
				return writer(c, val)
			}
		}
		return c.writeAvro(val)
	case mediaXML:
		if c.app != nil && c.app.hasWriters {
			if writer := c.app.writers[mediaXML]; writer != nil {
				return writer(c, val)
			}
		}
		return c.writeXML(val)
	default:
		if c.app != nil && c.app.hasWriters {
			if writer := c.app.writers[mediaJSON]; writer != nil {
				return writer(c, val)
			}
		}
		return c.writeJSON(val)
	}
}

func (c *Ctx) responseMediaType() mediaType {
	if c.acceptTypeCached {
		return c.acceptType
	}

	c.acceptType = acceptMediaType(c.Accept())
	c.acceptTypeCached = true
	return c.acceptType
}

func (c *Ctx) requestMediaType() mediaType {
	if c.contentTypeCached {
		return c.contentType
	}

	c.contentType = parseMediaType(c.requestContentType())
	c.contentTypeCached = true
	return c.contentType
}

func (c *Ctx) requestContentType() string {
	if c.contentTypeValueCached {
		return c.contentTypeValue
	}
	c.contentTypeValue = c.GetHeader("Content-Type")
	c.contentTypeValueCached = true
	return c.contentTypeValue
}

// writeJSON Write JSON
func (c *Ctx) writeJSON(val any) error {
	switch v := val.(type) {
	case nil:
		return nil
	case json.RawMessage:
		_, err := c.w.Write(v)
		return err
	default:
		return json.NewEncoder(c.w).Encode(val)
	}
}

// writeXML Write XML
func (c *Ctx) writeXML(val any) error {
	return xml.NewEncoder(c.w).Encode(val)
}

// writeGOB Write GOB
func (c *Ctx) writeGOB(val any) error {
	return gob.NewEncoder(c.w).Encode(val)
}

// writeBinary Write Binary
func (c *Ctx) writeBinary(val any) error {
	switch v := val.(type) {
	case nil:
		return nil
	case []byte:
		_, err := c.w.Write(v)
		return err
	case string:
		_, err := io.WriteString(c.w, v)
		return err
	case *bytes.Buffer:
		_, err := v.WriteTo(c.w)
		return err
	case io.Reader:
		buf := _copyBufPool.Get().(*[]byte)
		_, err := io.CopyBuffer(c.w, v, *buf)
		_copyBufPool.Put(buf)
		return err
	case encoding.BinaryMarshaler:
		b, err := v.MarshalBinary()
		if err != nil {
			return err
		}
		_, err = c.w.Write(b)
		return err
	default:
		return ErrNotImplemented
	}
}

// writeAvro Write Avro
func (c *Ctx) writeAvro(val any) error {
	switch v := val.(type) {
	case nil:
		return nil
	case AvroMarshaler:
		b, err := v.MarshalAvro()
		if err != nil {
			return err
		}
		_, err = c.w.Write(b)
		return err
	default:
		return c.writeBinary(val)
	}
}
