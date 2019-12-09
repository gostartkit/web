package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
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

// Validate decode and validate model
func (ctx *Context) Validate(v Validation) error {

	if err := json.NewDecoder(ctx.Request.Body).Decode(v); err != nil {
		return err
	}

	defer ctx.Request.Body.Close()

	return v.Validate(ctx.Request)
}

// Header get value by key from header
func (ctx *Context) Header(key string) string {
	return ctx.Request.Header.Get(key)
}

// Abort with http status and code
func (ctx *Context) Abort(status int, code int, err error) {
	ctx.ResponseWriter.WriteHeader(status)

	if err == nil {
		err = fmt.Errorf("Abort with %d", status)
	}

	data := createErrorResponse(code, err)
	ctx.WriteJSON(data)
	panic(data)
}

// Done with status
func (ctx *Context) Done(status int, result interface{}) {
	ctx.ResponseWriter.WriteHeader(status)

	data := createSuccessResponse(result)
	ctx.WriteJSON(data)
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
