package web

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
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
)

// createCtx returns a new instance of web.Ctx, initialized with the given HTTP response writer, request, and parameters.
func createCtx(w http.ResponseWriter, r *http.Request, params *Params) *Ctx {
	c := _ctxPool.Get().(*Ctx)
	c.w = w
	c.r = r
	c.param = params
	return c
}

// releaseCtx puts the context object back into the pool for reuse.
func releaseCtx(c *Ctx) {
	if c != nil {
		c.w = nil
		c.r = nil
		c.param = nil
		c.query = nil
		c.userId = 0
		c.accept = ""
		c.contentType = ""
		_ctxPool.Put(c)
	}
}

// Ctx represents the context for a web request, holding relevant request data and response methods.
type Ctx struct {
	w           http.ResponseWriter
	r           *http.Request
	param       *Params
	query       url.Values
	userId      uint64
	accept      string
	contentType string
}

// Init initializes the context with user ID and user rights.
func (c *Ctx) Init(userId uint64) {
	c.userId = userId
}

func (c *Ctx) Request() *http.Request {
	return c.r
}

func (c *Ctx) ResponseWriter() http.ResponseWriter {
	return c.w
}

func (c *Ctx) QueryValues() url.Values {
	if c.query == nil {
		c.query = c.r.URL.Query()
	}
	return c.query
}

// UserId returns the user id from the context.
func (c *Ctx) UserId() uint64 {
	return c.userId
}

// Param retrieves a parameter value by name from the Params.
func (c *Ctx) Param(name string) string {
	if c.param == nil {
		return ""
	}
	return c.param.Val(name)
}

// Query retrieves a query string parameter by name from the request URL.
func (c *Ctx) Query(name string) string {
	return c.QueryValues().Get(name)
}

// Form retrieves a form value by name from the request.
func (c *Ctx) Form(name string) string {
	return c.r.FormValue(name)
}

// PostForm retrieves a form value by name from the request.
func (c *Ctx) PostForm(name string) string {
	return c.r.PostFormValue(name)
}

// FormFile retrieves the first file uploaded for the specified form key.
// It calls Request.ParseMultipartForm and Request.ParseForm if needed.
func (c *Ctx) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.r.FormFile(key)
}

// Host returns the host from the request header.
func (c *Ctx) Host() string {
	return c.r.Host
}

// Path returns the path from the request URL.
func (c *Ctx) Path() string {
	return c.r.URL.Path
}

// Body returns the request body.
func (c *Ctx) Body() io.ReadCloser {
	return c.r.Body
}

// Method returns the HTTP method (GET, POST, etc.) used for the request.
func (c *Ctx) Method() string {
	return c.r.Method
}

// RemoteAddr returns the remote IP address of the client making the request.
func (c *Ctx) RemoteAddr() string {
	return c.r.RemoteAddr
}

// BearerToken retrieves the Bearer token from the Authorization header.
func (c *Ctx) BearerToken() string {
	return bearerToken(c.GetHeader("Authorization"))
}

// Origin returns the Origin header from the request.
func (c *Ctx) Origin() string {
	return c.GetHeader("Origin")
}

// SetOrigin sets the "Access-Control-Allow-Origin" header in the response.
func (c *Ctx) SetOrigin(origin string) {
	c.setHeader("Access-Control-Allow-Origin", origin)
}

// AllowCredentials sets the "Access-Control-Allow-Credentials" header to true in the response.
func (c *Ctx) AllowCredentials() {
	c.setHeader("Access-Control-Allow-Credentials", "true")
}

// UserAgent returns the User-Agent header from the request.
func (c *Ctx) UserAgent() string {
	return c.GetHeader("User-Agent")
}

// IsAjax checks if the request is an AJAX request based on the "X-Requested-With" header.
func (c *Ctx) IsAjax() bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

func (c *Ctx) IsFormData() bool {
	return strings.HasPrefix(c.ContentType(), "application/x-www-form-urlencoded") || strings.HasPrefix(c.ContentType(), "multipart/form-data")
}

// TryParseBody attempts to parse the request body based on its Content-Type and decode it into the provided value.
func (c *Ctx) TryParseBody(val any) error {

	if c.r == nil || c.r.Body == nil {
		return io.EOF
	}

	defer c.r.Body.Close()

	switch {
	case strings.HasPrefix(c.ContentType(), "application/json"):
		return json.NewDecoder(c.r.Body).Decode(val)
	case strings.HasPrefix(c.ContentType(), "application/x-gob"):
		return gob.NewDecoder(c.r.Body).Decode(val)
	case strings.HasPrefix(c.ContentType(), "application/octet-stream"):
		return ErrContentType
	case strings.HasPrefix(c.ContentType(), "application/xml"):
		return xml.NewDecoder(c.r.Body).Decode(val)
	default:
		return ErrContentType
	}
}

// TryParseParam attempts to parse a parameter value from the URL parameters.
func (c *Ctx) TryParseParam(name string, val any) error {
	return TryParse(c.Param(name), val)
}

// TryParseQuery attempts to parse a query string parameter value.
func (c *Ctx) TryParseQuery(name string, val any) error {
	return TryParse(c.Query(name), val)
}

// TryParseForm attempts to parse a form value by name.
func (c *Ctx) TryParseForm(name string, val any) error {
	return TryParse(c.Form(name), val)
}

// ParamInt attempts to parse an integer from the URL parameter by name.
func (c *Ctx) ParamInt(name string) (int, error) {
	return TryInt(c.Param(name))
}

// ParamUint attempts to parse an unsigned integer from the URL parameter by name.
func (c *Ctx) ParamUint(name string) (uint, error) {
	return TryUint(c.Param(name))
}

// ParamInt8 attempts to parse an int8 from the URL parameter by name.
func (c *Ctx) ParamInt8(name string) (int8, error) {
	return TryInt8(c.Param(name))
}

// ParamUint8 attempts to parse a uint8 from the URL parameter by name.
func (c *Ctx) ParamUint8(name string) (uint8, error) {
	return TryUint8(c.Param(name))
}

// ParamInt16 attempts to parse an int16 from the URL parameter by name.
func (c *Ctx) ParamInt16(name string) (int16, error) {
	return TryInt16(c.Param(name))
}

// ParamUint16 attempts to parse a uint16 from the URL parameter by name.
func (c *Ctx) ParamUint16(name string) (uint16, error) {
	return TryUint16(c.Param(name))
}

// ParamInt32 attempts to parse an int32 from the URL parameter by name.
func (c *Ctx) ParamInt32(name string) (int32, error) {
	return TryInt32(c.Param(name))
}

// ParamUint32 attempts to parse a uint32 from the URL parameter by name.
func (c *Ctx) ParamUint32(name string) (uint32, error) {
	return TryUint32(c.Param(name))
}

// ParamInt64 attempts to parse an int64 from the URL parameter by name.
func (c *Ctx) ParamInt64(name string) (int64, error) {
	return TryInt64(c.Param(name))
}

// ParamUint64 attempts to parse a uint64 from the URL parameter by name.
func (c *Ctx) ParamUint64(name string) (uint64, error) {
	return TryUint64(c.Param(name))
}

// ParamFloat32 attempts to parse a float32 from the URL parameter by name.
func (c *Ctx) ParamFloat32(name string) (float32, error) {
	return TryFloat32(c.Param(name))
}

// ParamFloat64 attempts to parse a float64 from the URL parameter by name.
func (c *Ctx) ParamFloat64(name string) (float64, error) {
	return TryFloat64(c.Param(name))
}

// ParamBool attempts to parse a bool from the URL parameter by name.
func (c *Ctx) ParamBool(name string) (bool, error) {
	return TryBool(c.Param(name))
}

// QueryInt attempts to parse a int from the request URL by name.
func (c *Ctx) QueryInt(name string) (int, error) {
	return TryInt(c.Query(name))
}

// QueryUint attempts to parse a uint from the request URL by name.
func (c *Ctx) QueryUint(name string) (uint, error) {
	return TryUint(c.Query(name))
}

// QueryInt8 attempts to parse a int8 from the request URL by name.
func (c *Ctx) QueryInt8(name string) (int8, error) {
	return TryInt8(c.Query(name))
}

// QueryUint8 attempts to parse a uint8 from the request URL by name.
func (c *Ctx) QueryUint8(name string) (uint8, error) {
	return TryUint8(c.Query(name))
}

// QueryInt16 attempts to parse a int16 from the request URL by name.
func (c *Ctx) QueryInt16(name string) (int16, error) {
	return TryInt16(c.Query(name))
}

// QueryUint16 attempts to parse a uint16 from the request URL by name.
func (c *Ctx) QueryUint16(name string) (uint16, error) {
	return TryUint16(c.Query(name))
}

// QueryInt32 attempts to parse a int32 from the request URL by name.
func (c *Ctx) QueryInt32(name string) (int32, error) {
	return TryInt32(c.Query(name))
}

// QueryUint32 attempts to parse a uint32 from the request URL by name.
func (c *Ctx) QueryUint32(name string) (uint32, error) {
	return TryUint32(c.Query(name))
}

// QueryInt64 attempts to parse a int64 from the request URL by name.
func (c *Ctx) QueryInt64(name string) (int64, error) {
	return TryInt64(c.Query(name))
}

// QueryUint64 attempts to parse a uint64 from the request URL by name.
func (c *Ctx) QueryUint64(name string) (uint64, error) {
	return TryUint64(c.Query(name))
}

// QueryFloat32 attempts to parse a float32 from the request URL by name.
func (c *Ctx) QueryFloat32(name string) (float32, error) {
	return TryFloat32(c.Query(name))
}

// QueryFloat64 attempts to parse a float64 from the request URL by name.
func (c *Ctx) QueryFloat64(name string) (float64, error) {
	return TryFloat64(c.Query(name))
}

// QueryBool attempts to parse a bool from the request URL by name.
func (c *Ctx) QueryBool(name string) (bool, error) {
	return TryBool(c.Query(name))
}

// FormIn attempts to parse a int from form value by name.
func (c *Ctx) FormInt(name string) (int, error) {
	return TryInt(c.Form(name))
}

// FormUint attempts to parse a uint from form value by name.
func (c *Ctx) FormUint(name string) (uint, error) {
	return TryUint(c.Form(name))
}

// FormInt8 attempts to parse a int8 from form value by name.
func (c *Ctx) FormInt8(name string) (int8, error) {
	return TryInt8(c.Form(name))
}

// FormUint8 attempts to parse a uint8 from form value by name.
func (c *Ctx) FormUint8(name string) (uint8, error) {
	return TryUint8(c.Form(name))
}

// FormInt16 attempts to parse a int16 from form value by name.
func (c *Ctx) FormInt16(name string) (int16, error) {
	return TryInt16(c.Form(name))
}

// FormUint16 attempts to parse a uint16 from form value by name.
func (c *Ctx) FormUint16(name string) (uint16, error) {
	return TryUint16(c.Form(name))
}

// FormInt32 attempts to parse a int32 from form value by name.
func (c *Ctx) FormInt32(name string) (int32, error) {
	return TryInt32(c.Form(name))
}

// FormUint32 attempts to parse a uint32 from form value by name.
func (c *Ctx) FormUint32(name string) (uint32, error) {
	return TryUint32(c.Form(name))
}

// FormInt64 attempts to parse a int64 from form value by name.
func (c *Ctx) FormInt64(name string) (int64, error) {
	return TryInt64(c.Form(name))
}

// FormUint64 attempts to parse a uint64 from form value by name.
func (c *Ctx) FormUint64(name string) (uint64, error) {
	return TryUint64(c.Form(name))
}

// FormFloat32 attempts to parse a float32 from form value by name.
func (c *Ctx) FormFloat32(name string) (float32, error) {
	return TryFloat32(c.Form(name))
}

// FormFloat64 attempts to parse a float64 from form value by name.
func (c *Ctx) FormFloat64(name string) (float64, error) {
	return TryFloat64(c.Form(name))
}

// FormBool attempts to parse a bool from form value by name.
func (c *Ctx) FormBool(name string) (bool, error) {
	return TryBool(c.Form(name))
}

// PostFormIn attempts to parse a int from PostForm value by name.
func (c *Ctx) PostFormInt(name string) (int, error) {
	return TryInt(c.PostForm(name))
}

// PostFormUint attempts to parse a uint from PostForm value by name.
func (c *Ctx) PostFormUint(name string) (uint, error) {
	return TryUint(c.PostForm(name))
}

// PostFormInt8 attempts to parse a int8 from PostForm value by name.
func (c *Ctx) PostFormInt8(name string) (int8, error) {
	return TryInt8(c.PostForm(name))
}

// PostFormUint8 attempts to parse a uint8 from PostForm value by name.
func (c *Ctx) PostFormUint8(name string) (uint8, error) {
	return TryUint8(c.PostForm(name))
}

// PostFormInt16 attempts to parse a int16 from PostForm value by name.
func (c *Ctx) PostFormInt16(name string) (int16, error) {
	return TryInt16(c.PostForm(name))
}

// PostFormUint16 attempts to parse a uint16 from PostForm value by name.
func (c *Ctx) PostFormUint16(name string) (uint16, error) {
	return TryUint16(c.PostForm(name))
}

// PostFormInt32 attempts to parse a int32 from PostForm value by name.
func (c *Ctx) PostFormInt32(name string) (int32, error) {
	return TryInt32(c.PostForm(name))
}

// PostFormUint32 attempts to parse a uint32 from PostForm value by name.
func (c *Ctx) PostFormUint32(name string) (uint32, error) {
	return TryUint32(c.PostForm(name))
}

// PostFormInt64 attempts to parse a int64 from PostForm value by name.
func (c *Ctx) PostFormInt64(name string) (int64, error) {
	return TryInt64(c.PostForm(name))
}

// PostFormUint64 attempts to parse a uint64 from PostForm value by name.
func (c *Ctx) PostFormUint64(name string) (uint64, error) {
	return TryUint64(c.PostForm(name))
}

// PostFormFloat32 attempts to parse a float32 from PostForm value by name.
func (c *Ctx) PostFormFloat32(name string) (float32, error) {
	return TryFloat32(c.PostForm(name))
}

// PostFormFloat64 attempts to parse a float64 from PostForm value by name.
func (c *Ctx) PostFormFloat64(name string) (float64, error) {
	return TryFloat64(c.PostForm(name))
}

// PostFormBool attempts to parse a bool from PostForm value by name.
func (c *Ctx) PostFormBool(name string) (bool, error) {
	return TryBool(c.PostForm(name))
}

// Accept get Accept from header
func (c *Ctx) Accept() string {
	if c.accept == "" {
		c.accept = c.GetHeader("Accept")
	}
	return c.accept
}

// Flusher returns the http.Flusher interface if the response writer supports it.
// This is useful for enabling HTTP/1.1 chunked transfer encoding.
// It allows the server to send data to the client in chunks, rather than waiting for the entire response to be ready.
// This is particularly useful for streaming data or for long-lived connections.
// If the response writer does not support chunked transfer encoding, it returns nil.
func (c *Ctx) Flusher() http.Flusher {
	if flusher, ok := c.w.(http.Flusher); ok {
		return flusher
	}
	return nil
}

// Hijacker returns the http.Hijacker interface if the response writer supports it.
// This is useful for upgrading the connection to a different protocol, such as WebSocket.
// If the response writer does not support hijacking, it returns nil.
func (c *Ctx) Hijacker() http.Hijacker {
	if hijacker, ok := c.w.(http.Hijacker); ok {
		return hijacker
	}
	return nil
}

// Context returns the context of the request.
func (c *Ctx) Context() context.Context {
	return c.r.Context()
}

// ContentType get Content-Type from header
func (c *Ctx) ContentType() string {
	if c.contentType == "" {
		c.contentType = c.GetHeader("Content-Type")
	}
	return c.contentType
}

// SetContentType Set Content-Type to header
func (c *Ctx) SetContentType(val string) {
	c.contentType = val
	c.setHeader("Content-Type", val)
}

// SetCacheControl Set Cache-Control to header
func (c *Ctx) SetCacheControl(val string) {
	c.setHeader("Cache-Control", val)
}

// SetConnection Set Connection to header
func (c *Ctx) SetConnection(val string) {
	c.setHeader("Connection", val)
}

// SetVersion set `version` header
func (c *Ctx) SetVersion(version string) {
	c.setHeader("Version", version)
}

// SetCookie adds a Set-Cookie header to the provided [ResponseWriter]'s headers. The provided cookie must have a valid Name. Invalid cookies may be silently dropped.
func (c *Ctx) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.w, cookie)
}

// GetCookie returns the named cookie provided in the request or [ErrNoCookie] if not found. If multiple cookies match the given name, only one cookie will be returned.
func (c *Ctx) GetCookie(name string) (*http.Cookie, error) {
	return c.r.Cookie(name)
}

// GetHeader GetHeader header, short hand of r.Header.GetHeader
func (c *Ctx) GetHeader(key string) string {
	return c.r.Header.Get(key)
}

// setHeader setHeader header, short hand of w.Header().setHeader
func (c *Ctx) setHeader(key string, value string) {
	c.w.Header().Set(key, value)
}

// write write data base on accept header
func (c *Ctx) write(val any) error {

	switch c.Accept() {
	case "application/json":
		return c.writeJSON(val)
	case "application/x-gob":
		return c.writeGOB(val)
	case "application/octet-stream":
		return c.writeBinary(val)
	case "application/x-avro":
		return c.writeAvro(val)
	case "application/xml":
		return c.writeXML(val)
	default:
		return c.writeJSON(val)
	}
}

// writeJSON Write JSON
func (c *Ctx) writeJSON(val any) error {
	return json.NewEncoder(c.w).Encode(val)
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
	return ErrNotImplemented
}

// writeAvro Write Avro
func (c *Ctx) writeAvro(val any) error {
	return ErrNotImplemented
}
