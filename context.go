package web

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
)

// createContext return a web.Context
func createContext(w http.ResponseWriter, r *http.Request, params *Params) *Context {

	ctx := &Context{
		ResponseWriter: w,
		Request:        r,
		params:         params,
	}

	return ctx
}

// Context is type of an web.Context
type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	params         *Params
	urlValues      *url.Values
	UserID         uint64
}

// Param get value from Params
func (ctx *Context) Param(name string) string {
	return ctx.params.Val(name)
}

// Query get value from QueryString
func (ctx *Context) Query(name string) string {
	if ctx.urlValues == nil {
		urlValues := ctx.Request.URL.Query()
		ctx.urlValues = &urlValues
	}

	return ctx.urlValues.Get(name)
}

// Form get value from Form
func (ctx *Context) Form(name string) string {
	if ctx.Request.Form == nil {
		ctx.Request.ParseForm()
	}
	return ctx.Request.Form.Get(name)
}

// TryParseBody decode val from Request.Body
func (ctx *Context) TryParseBody(val interface{}) error {
	if err := json.NewDecoder(ctx.Request.Body).Decode(val); err != nil {
		return err
	}
	defer ctx.Request.Body.Close()
	return nil
}

// TryParseParam decode val from Query
func (ctx *Context) TryParseParam(name string, val interface{}) error {
	return tryParse(ctx.Param(name), val)
}

// TryParseQuery decode val from Query
func (ctx *Context) TryParseQuery(name string, val interface{}) error {
	return tryParse(ctx.Query(name), val)
}

// TryParseForm decode val from Form
func (ctx *Context) TryParseForm(name string, val interface{}) error {
	return tryParse(ctx.Form(name), val)
}

// GetHeader get header by key
func (ctx *Context) GetHeader(key string) string {
	return ctx.Request.Header.Get(key)
}

// WriteBytes Write bytes
func (ctx *Context) WriteBytes(val []byte) (int, error) {
	return ctx.ResponseWriter.Write(val)
}

// Bytes Write bytes with status code
func (ctx *Context) Bytes(code int, val []byte) (int, error) {
	ctx.ResponseWriter.WriteHeader(code)
	return ctx.ResponseWriter.Write(val)
}

// WriteString Write String
func (ctx *Context) WriteString(val string) (int, error) {
	return ctx.ResponseWriter.Write([]byte(val))
}

// String Write String with status code
func (ctx *Context) String(code int, val string) (int, error) {
	ctx.ResponseWriter.WriteHeader(code)
	return ctx.ResponseWriter.Write([]byte(val))
}

// WriteJSON Write JSON
func (ctx *Context) WriteJSON(val interface{}) error {
	return json.NewEncoder(ctx.ResponseWriter).Encode(val)
}

// JSON Write JSON with status code
func (ctx *Context) JSON(code int, val interface{}) error {
	ctx.ResponseWriter.WriteHeader(code)
	return json.NewEncoder(ctx.ResponseWriter).Encode(val)
}

// WriteXML Write XML
func (ctx *Context) WriteXML(val interface{}) error {
	return xml.NewEncoder(ctx.ResponseWriter).Encode(val)
}

// XML Write XML with status code
func (ctx *Context) XML(code int, val interface{}) error {
	ctx.ResponseWriter.WriteHeader(code)
	return xml.NewEncoder(ctx.ResponseWriter).Encode(val)
}

// Status Write status code to header
func (ctx *Context) Status(code int) {
	ctx.ResponseWriter.WriteHeader(code)
}

// SetHeader Set key/value to header
func (ctx *Context) SetHeader(key string, value string) {
	ctx.ResponseWriter.Header().Set(key, value)
}

// AddHeader Add key/value to header
func (ctx *Context) AddHeader(key string, value string) {
	ctx.ResponseWriter.Header().Add(key, value)
}

// GetContentType get Content-Type from header
func (ctx *Context) GetContentType() string {
	return ctx.GetHeader("Content-Type")
}

// SetContentType Set Content-Type to header
func (ctx *Context) SetContentType(val string) {
	ctx.SetHeader("Content-Type", contentType(val))
}

// Redirect to url with status code
func (ctx *Context) Redirect(code int, url string) {
	ctx.SetHeader("Location", url)
	ctx.ResponseWriter.WriteHeader(code)
	ctx.WriteString("Redirecting to: " + url)
}
