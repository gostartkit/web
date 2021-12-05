package web

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// createContext return a web.Context
func createContext(w http.ResponseWriter, r *http.Request, params *Params) *Context {

	ctx := &Context{
		W:     w,
		R:     r,
		param: params,
		code:  200,
	}

	return ctx
}

// Context is type of an web.Context
type Context struct {
	W http.ResponseWriter
	R *http.Request

	param       *Params
	query       *url.Values
	form        *url.Values
	userID      uint64
	accept      *string
	contentType *string
	code        int
}

// Init init context
func (ctx *Context) Init(userID uint64) {
	ctx.userID = userID
}

// UserID get userID
func (ctx *Context) UserID() uint64 {
	return ctx.userID
}

// Param get value from Params
func (ctx *Context) Param(name string) string {
	return ctx.param.Val(name)
}

// Query get value from QueryString
func (ctx *Context) Query(name string) string {
	if ctx.query == nil {
		query := ctx.R.URL.Query()
		ctx.query = &query
	}
	return ctx.query.Get(name)
}

// Form get value from Form
func (ctx *Context) Form(name string) string {
	if ctx.form == nil {
		ctx.form, _ = parseForm(ctx.R.Body)
	}
	return ctx.form.Get(name)
}

// Host return ctx.r.Host
func (ctx *Context) Host() string {
	return ctx.R.Host
}

// Path return ctx.r.URL.Path
func (ctx *Context) Path() string {
	return ctx.R.URL.Path
}

// Method return ctx.r.Method
func (ctx *Context) Method() string {
	return ctx.R.Method
}

// RemoteAddr return remote ip address
func (ctx *Context) RemoteAddr() string {
	return ctx.R.RemoteAddr
}

// UserAgent return User-Agent header
func (ctx *Context) UserAgent() string {
	return ctx.Get("User-Agent")
}

// IsAjax if X-Requested-With header is XMLHttpRequest return true, else false
func (ctx *Context) IsAjax() bool {
	return ctx.Get("X-Requested-With") == "XMLHttpRequest"
}

// TryParseBody decode val from Request.Body
func (ctx *Context) TryParseBody(val interface{}) error {
	switch {
	case strings.HasPrefix(ctx.ContentType(), "application/json"):
		return json.NewDecoder(ctx.R.Body).Decode(val)
	case strings.HasPrefix(ctx.ContentType(), "application/x-gob"):
		return gob.NewDecoder(ctx.R.Body).Decode(val)
	case strings.HasPrefix(ctx.ContentType(), "application/x-www-form-urlencoded"):
		return formReader(ctx, val)
	case strings.HasPrefix(ctx.ContentType(), "multipart/form-data"):
		return formDataReader(ctx, val)
	case strings.HasPrefix(ctx.ContentType(), "application/octet-stream"):
		return binaryReader(ctx, val)
	case strings.HasPrefix(ctx.ContentType(), "application/xml"):
		return xml.NewDecoder(ctx.R.Body).Decode(val)
	default:
		return errors.New("tryParseBody(unsupported contentType '" + ctx.ContentType() + "')")
	}
}

// TryParseParam decode val from Query
func (ctx *Context) TryParseParam(name string, val interface{}) error {
	return TryParse(ctx.Param(name), val)
}

// TryParseQuery decode val from Query
func (ctx *Context) TryParseQuery(name string, val interface{}) error {
	return TryParse(ctx.Query(name), val)
}

// TryParseForm decode val from Form
func (ctx *Context) TryParseForm(name string, val interface{}) error {
	return TryParse(ctx.Form(name), val)
}

// Write Write data base on accept header
func (ctx *Context) Write(val interface{}) error {
	switch ctx.Accept() {
	case "application/octet-stream", "application/x-avro":
		return ctx.WriteBinary(val)
	case "application/x-gob":
		return ctx.WriteGOB(val)
	case "application/xml":
		return ctx.WriteXML(val)
	default:
		return ctx.WriteJSON(val)
	}
}

// WriteJSON Write JSON
func (ctx *Context) WriteJSON(val interface{}) error {
	ctx.W.WriteHeader(ctx.code)
	return json.NewEncoder(ctx.W).Encode(val)
}

// WriteXML Write XML
func (ctx *Context) WriteXML(val interface{}) error {
	ctx.W.WriteHeader(ctx.code)
	return xml.NewEncoder(ctx.W).Encode(val)
}

// WriteGOB Write GOB
func (ctx *Context) WriteGOB(val interface{}) error {
	ctx.W.WriteHeader(ctx.code)
	return gob.NewEncoder(ctx.W).Encode(val)
}

// WriteBinary Write Binary
func (ctx *Context) WriteBinary(val interface{}) error {
	return binaryWriter(ctx, val)
}

// Status return status code
func (ctx *Context) Status() int {
	return ctx.code
}

// SetStatus Write status code to header
func (ctx *Context) SetStatus(code int) {
	ctx.code = code
}

// Get get header, short hand for ctx.Request.Header.Get
func (ctx *Context) Get(key string) string {
	return ctx.R.Header.Get(key)
}

// Set set header, short hand for ctx.ResponseWriter.Header().Set
func (ctx *Context) Set(key string, value string) {
	ctx.W.Header().Set(key, value)
}

// Add add header, short hand for ctx.ResponseWriter.Header().Add
func (ctx *Context) Add(key string, value string) {
	ctx.W.Header().Add(key, value)
}

// Del del header, short hand for ctx.ResponseWriter.Header().Del
func (ctx *Context) Del(key string) {
	ctx.W.Header().Del(key)
}

// Accept get Accept from header
func (ctx *Context) Accept() string {
	if ctx.accept == nil {
		ac := ctx.Get("Accept")
		ctx.accept = &ac
	}
	return *ctx.accept
}

// ContentType get Content-Type from header
func (ctx *Context) ContentType() string {
	if ctx.contentType == nil {
		ctype := ctx.Get("Content-Type")
		ctx.contentType = &ctype
	}
	return *ctx.contentType
}

// SetContentType Set Content-Type to header
func (ctx *Context) SetContentType(val string) {
	ctx.Set("Content-Type", contentType(val))
}

// Redirect to url with status code
func (ctx *Context) Redirect(code int, url string) {
	ctx.Set("Location", url)
	ctx.SetStatus(code)
	ctx.W.WriteHeader(code)
}

// WriteContentType write content type to client
func (ctx *Context) WriteContentType() {

	ac := ctx.Accept()

	switch ac {
	case "application/octet-stream", "application/x-avro", "application/x-gob", "application/xml":
		ctx.SetContentType(ac)
	default:
		ctx.SetContentType("application/json")
	}
}
