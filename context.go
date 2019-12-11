package web

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
)

// Context is type of an web.Context
type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	paramValues    *Params
	queryValues    url.Values
	formValues     url.Values
}

// Val get value from Params
func (ctx *Context) Val(name string) string {
	return ctx.paramValues.Val(name)
}

// Query get value from QueryString
func (ctx *Context) Query(name string) string {
	if ctx.queryValues == nil {
		ctx.queryValues = ctx.Request.URL.Query()
	}

	return ctx.queryValues.Get(name)
}

// Form get value from Form
func (ctx *Context) Form(name string) string {
	if ctx.formValues == nil {
		if ctx.Request.Form == nil {
			ctx.Request.ParseForm()
		}

		ctx.formValues = ctx.Request.Form
	}

	return ctx.formValues.Get(name)
}

// Parse decode val from Request.Body
func (ctx *Context) Parse(val interface{}) error {

	if err := json.NewDecoder(ctx.Request.Body).Decode(val); err != nil {
		return err
	}

	defer ctx.Request.Body.Close()

	return nil
}

// ParseAbortIf decode val from Request.Body
func (ctx *Context) ParseAbortIf(val interface{}) {
	ctx.AbortIf(400, ctx.Parse(val), nil)
}

// AbortIf with error
func (ctx *Context) AbortIf(code int, v interface{}, fn func(code int, v interface{}) error) {
	if v != nil {
		if fn != nil {
			fn(code, v)
		}
		panic(v)
	}
}

// Abort with error
func (ctx *Context) Abort(v interface{}) {
	panic(v)
}

// Header get value by key from header
func (ctx *Context) Header(key string) string {
	return ctx.Request.Header.Get(key)
}

// Write bytes
func (ctx *Context) Write(val []byte) (int, error) {
	return ctx.ResponseWriter.Write(val)
}

// WriteString Write String
func (ctx *Context) WriteString(val string) (int, error) {
	return ctx.ResponseWriter.Write([]byte(val))
}

// WriteJSON Write JSON
func (ctx *Context) WriteJSON(val interface{}) error {
	return json.NewEncoder(ctx.ResponseWriter).Encode(val)
}

// WriteXML Write XML
func (ctx *Context) WriteXML(val interface{}) error {
	return xml.NewEncoder(ctx.ResponseWriter).Encode(val)
}

// WriteSuccess with status
func (ctx *Context) WriteSuccess(code int, result interface{}) error {
	data := &responseData{
		Success: true,
		Code:    code,
		Result:  result,
	}
	ctx.ResponseWriter.WriteHeader(200)
	return ctx.WriteJSON(data)
}

// WriteError with http 400 and code
func (ctx *Context) WriteError(code int, err error) error {
	data := &responseData{
		Success: false,
		Code:    code,
		Error:   err,
	}
	ctx.ResponseWriter.WriteHeader(400)
	return ctx.WriteJSON(data)
}

// WriteHeader Write Header
func (ctx *Context) WriteHeader(statusCode int) {
	ctx.ResponseWriter.WriteHeader(statusCode)
}

// SetHeader Set Header
func (ctx *Context) SetHeader(key string, value string) {
	ctx.ResponseWriter.Header().Set(key, value)
}

// AddHeader Add Header
func (ctx *Context) AddHeader(key string, value string) {
	ctx.ResponseWriter.Header().Add(key, value)
}

// SetContentType Set Content-Type
func (ctx *Context) SetContentType(val string) {
	ctx.ResponseWriter.Header().Set("Content-Type", contentType(val))
}

// Redirect to url with status
func (ctx *Context) Redirect(status int, url string) {
	ctx.SetHeader("Location", url)
	ctx.WriteHeader(status)
	ctx.WriteString("Redirecting to: " + url)
}
