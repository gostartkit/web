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
	"strconv"
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
	c.query = nil
	c.userID = 0
	c.accept = nil
	c.contentType = nil

	return c
}

// releaseCtx puts the context object back into the pool for reuse.
func releaseCtx(c *Ctx) {
	if c != nil {
		_ctxPool.Put(c)
	}
}

// Ctx represents the context for a web request, holding relevant request data and response methods.
type Ctx struct {
	w           http.ResponseWriter
	r           *http.Request
	param       *Params
	query       *url.Values
	userID      uint64
	accept      *string
	contentType *string
}

// Init initializes the context with user ID and user rights.
func (c *Ctx) Init(userID uint64) {
	c.userID = userID
}

// UserID returns the user ID from the context.
func (c *Ctx) UserID() uint64 {
	return c.userID
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
	if c.query == nil {
		query := c.r.URL.Query()
		c.query = &query
	}
	return c.query.Get(name)
}

// Form retrieves a form value by name from the request.
func (c *Ctx) Form(name string) string {
	return c.r.FormValue(name)
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
	return bearerToken(c.Get("Authorization"))
}

// Origin returns the Origin header from the request.
func (c *Ctx) Origin() string {
	return c.Get("Origin")
}

// SetOrigin sets the "Access-Control-Allow-Origin" header in the response.
func (c *Ctx) SetOrigin(origin string) {
	c.set("Access-Control-Allow-Origin", origin)
}

// AllowCredentials sets the "Access-Control-Allow-Credentials" header to true in the response.
func (c *Ctx) AllowCredentials() {
	c.set("Access-Control-Allow-Credentials", "true")
}

// UserAgent returns the User-Agent header from the request.
func (c *Ctx) UserAgent() string {
	return c.Get("User-Agent")
}

// IsAjax checks if the request is an AJAX request based on the "X-Requested-With" header.
func (c *Ctx) IsAjax() bool {
	return c.Get("X-Requested-With") == "XMLHttpRequest"
}

// TryParseBody attempts to parse the request body based on its Content-Type and decode it into the provided value.
func (c *Ctx) TryParseBody(val any) error {
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
	n, err := strconv.ParseInt(c.Param(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// ParamUint attempts to parse an unsigned integer from the URL parameter by name.
func (c *Ctx) ParamUint(name string) (uint, error) {
	n, err := strconv.ParseUint(c.Param(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// ParamInt8 attempts to parse an int8 from the URL parameter by name.
func (c *Ctx) ParamInt8(name string) (int8, error) {
	n, err := strconv.ParseInt(c.Param(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(n), nil
}

// ParamUint8 attempts to parse a uint8 from the URL parameter by name.
func (c *Ctx) ParamUint8(name string) (uint8, error) {
	n, err := strconv.ParseUint(c.Param(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(n), nil
}

// ParamInt16 attempts to parse an int16 from the URL parameter by name.
func (c *Ctx) ParamInt16(name string) (int16, error) {
	n, err := strconv.ParseInt(c.Param(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(n), nil
}

// ParamUint16 attempts to parse a uint16 from the URL parameter by name.
func (c *Ctx) ParamUint16(name string) (uint16, error) {
	n, err := strconv.ParseUint(c.Param(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), nil
}

// ParamInt32 attempts to parse an int32 from the URL parameter by name.
func (c *Ctx) ParamInt32(name string) (int32, error) {
	n, err := strconv.ParseInt(c.Param(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}

// ParamUint32 attempts to parse a uint32 from the URL parameter by name.
func (c *Ctx) ParamUint32(name string) (uint32, error) {
	n, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// ParamInt64 attempts to parse an int64 from the URL parameter by name.
func (c *Ctx) ParamInt64(name string) (int64, error) {
	n, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ParamUint64 attempts to parse a uint64 from the URL parameter by name.
func (c *Ctx) ParamUint64(name string) (uint64, error) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ParamFloat32 attempts to parse a float32 from the URL parameter by name.
func (c *Ctx) ParamFloat32(name string) (float32, error) {
	n, err := strconv.ParseFloat(c.Param(name), 32)
	if err != nil {
		return 0, err
	}
	return float32(n), nil
}

// ParamFloat64 attempts to parse a float64 from the URL parameter by name.
func (c *Ctx) ParamFloat64(name string) (float64, error) {
	n, err := strconv.ParseFloat(c.Param(name), 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ParamBool attempts to parse a bool from the URL parameter by name.
func (c *Ctx) ParamBool(name string) (bool, error) {
	n, err := strconv.ParseBool(c.Param(name))
	if err != nil {
		return false, err
	}
	return n, nil
}

// QueryInt attempts to parse a int from the request URL by name.
func (c *Ctx) QueryInt(name string) (int, error) {
	n, err := strconv.ParseInt(c.Query(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// QueryUint attempts to parse a uint from the request URL by name.
func (c *Ctx) QueryUint(name string) (uint, error) {
	n, err := strconv.ParseUint(c.Query(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// QueryInt8 attempts to parse a int8 from the request URL by name.
func (c *Ctx) QueryInt8(name string) (int8, error) {
	n, err := strconv.ParseInt(c.Query(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(n), nil
}

// QueryUint8 attempts to parse a uint8 from the request URL by name.
func (c *Ctx) QueryUint8(name string) (uint8, error) {
	n, err := strconv.ParseUint(c.Query(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(n), nil
}

// QueryInt16 attempts to parse a int16 from the request URL by name.
func (c *Ctx) QueryInt16(name string) (int16, error) {
	n, err := strconv.ParseInt(c.Query(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(n), nil
}

// QueryUint16 attempts to parse a uint16 from the request URL by name.
func (c *Ctx) QueryUint16(name string) (uint16, error) {
	n, err := strconv.ParseUint(c.Query(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), nil
}

// QueryInt32 attempts to parse a int32 from the request URL by name.
func (c *Ctx) QueryInt32(name string) (int32, error) {
	n, err := strconv.ParseInt(c.Query(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}

// QueryUint32 attempts to parse a uint32 from the request URL by name.
func (c *Ctx) QueryUint32(name string) (uint32, error) {
	n, err := strconv.ParseUint(c.Query(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// QueryInt64 attempts to parse a int64 from the request URL by name.
func (c *Ctx) QueryInt64(name string) (int64, error) {
	n, err := strconv.ParseInt(c.Query(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// QueryUint64 attempts to parse a uint64 from the request URL by name.
func (c *Ctx) QueryUint64(name string) (uint64, error) {
	n, err := strconv.ParseUint(c.Query(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// QueryFloat32 attempts to parse a float32 from the request URL by name.
func (c *Ctx) QueryFloat32(name string) (float32, error) {
	n, err := strconv.ParseFloat(c.Query(name), 32)
	if err != nil {
		return 0, err
	}
	return float32(n), nil
}

// QueryFloat64 attempts to parse a float64 from the request URL by name.
func (c *Ctx) QueryFloat64(name string) (float64, error) {
	n, err := strconv.ParseFloat(c.Query(name), 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// QueryBool attempts to parse a bool from the request URL by name.
func (c *Ctx) QueryBool(name string) (bool, error) {
	n, err := strconv.ParseBool(c.Query(name))
	if err != nil {
		return false, err
	}
	return n, nil
}

// FormIn attempts to parse a int from form value by name.
func (c *Ctx) FormInt(name string) (int, error) {
	n, err := strconv.ParseInt(c.Form(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// FormUint attempts to parse a uint from form value by name.
func (c *Ctx) FormUint(name string) (uint, error) {
	n, err := strconv.ParseUint(c.Form(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// FormInt8 attempts to parse a int8 from form value by name.
func (c *Ctx) FormInt8(name string) (int8, error) {
	n, err := strconv.ParseInt(c.Form(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(n), nil
}

// FormUint8 attempts to parse a uint8 from form value by name.
func (c *Ctx) FormUint8(name string) (uint8, error) {
	n, err := strconv.ParseUint(c.Form(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(n), nil
}

// FormInt16 attempts to parse a int16 from form value by name.
func (c *Ctx) FormInt16(name string) (int16, error) {
	n, err := strconv.ParseInt(c.Form(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(n), nil
}

// FormUint16 attempts to parse a uint16 from form value by name.
func (c *Ctx) FormUint16(name string) (uint16, error) {
	n, err := strconv.ParseUint(c.Form(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), nil
}

// FormInt32 attempts to parse a int32 from form value by name.
func (c *Ctx) FormInt32(name string) (int32, error) {
	n, err := strconv.ParseInt(c.Form(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}

// FormUint32 attempts to parse a uint32 from form value by name.
func (c *Ctx) FormUint32(name string) (uint32, error) {
	n, err := strconv.ParseUint(c.Form(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// FormInt64 attempts to parse a int64 from form value by name.
func (c *Ctx) FormInt64(name string) (int64, error) {
	n, err := strconv.ParseInt(c.Form(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// FormUint64 attempts to parse a uint64 from form value by name.
func (c *Ctx) FormUint64(name string) (uint64, error) {
	n, err := strconv.ParseUint(c.Form(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// FormFloat32 attempts to parse a float32 from form value by name.
func (c *Ctx) FormFloat32(name string) (float32, error) {
	n, err := strconv.ParseFloat(c.Form(name), 32)
	if err != nil {
		return 0, err
	}
	return float32(n), nil
}

// FormFloat64 attempts to parse a float64 from form value by name.
func (c *Ctx) FormFloat64(name string) (float64, error) {
	n, err := strconv.ParseFloat(c.Form(name), 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// FormBool attempts to parse a bool from form value by name.
func (c *Ctx) FormBool(name string) (bool, error) {
	n, err := strconv.ParseBool(c.Form(name))
	if err != nil {
		return false, err
	}
	return n, nil
}

// QueryFilter c.Query(QueryFilter)
func (c *Ctx) QueryFilter() string {
	return c.Query(QueryFilter)
}

// QueryOrderBy c.Query(QueryOrderBy)
func (c *Ctx) QueryOrderBy() string {
	return c.Query(QueryOrderBy)
}

// QueryPage c.QueryInt(QueryPage)
func (c *Ctx) QueryPage(defaultPage int) int {

	page, err := c.QueryInt(QueryPage)

	if err != nil {
		page = defaultPage
	}

	return page
}

// QueryPageSize c.QueryInt(QueryPageSize)
func (c *Ctx) QueryPageSize(defaultPageSize int) int {

	pageSize, err := c.QueryInt(QueryPageSize)

	if err != nil {
		pageSize = defaultPageSize
	}

	return pageSize
}

// HeaderAttrs strings.Split(c.Get(HeaderAttrs), ",")
func (c *Ctx) HeaderAttrs() []string {

	attrs := c.Get(HeaderAttrs)

	return strings.Split(attrs, ",")
}

// Accept get Accept from header
func (c *Ctx) Accept() string {
	if c.accept == nil {
		ac := c.Get("Accept")
		c.accept = &ac
	}
	return *c.accept
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
	if c.contentType == nil {
		ctype := c.Get("Content-Type")
		c.contentType = &ctype
	}
	return *c.contentType
}

// SetContentType Set Content-Type to header
func (c *Ctx) SetContentType(val string) {
	if c.contentType == nil {
		c.contentType = &val
	}
	c.set("Content-Type", val)
}

// SetCacheControl Set Cache-Control to header
func (c *Ctx) SetCacheControl(val string) {
	c.set("Cache-Control", val)
}

// SetConnection Set Connection to header
func (c *Ctx) SetConnection(val string) {
	c.set("Connection", val)
}

// SetVersion set `version` header
func (c *Ctx) SetVersion(version string) {
	c.set("Version", version)
}

// SetCookie adds a Set-Cookie header to the provided [ResponseWriter]'s headers. The provided cookie must have a valid Name. Invalid cookies may be silently dropped.
func (c *Ctx) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.w, cookie)
}

// Cookie returns the named cookie provided in the request or [ErrNoCookie] if not found. If multiple cookies match the given name, only one cookie will be returned.
func (c *Ctx) Cookie(name string) (*http.Cookie, error) {
	return c.r.Cookie(name)
}

// Get Get header, short hand of r.Header.Get
func (c *Ctx) Get(key string) string {
	return c.r.Header.Get(key)
}

// set set header, short hand of w.Header().set
func (c *Ctx) set(key string, value string) {
	c.w.Header().Set(key, value)
}

// write write data base on accept header
func (c *Ctx) write(val any) error {

	switch c.ContentType() {
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
