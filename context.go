package web

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
)

// Context is type of an web.Context
type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	params         *Params
	urlValues      *url.Values
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

// TryParse decode val from Request.Body
func (ctx *Context) TryParse(val interface{}) error {
	if err := json.NewDecoder(ctx.Request.Body).Decode(val); err != nil {
		return err
	}
	defer ctx.Request.Body.Close()
	return nil
}

// Parse decode val from Request.Body, if error != nil abort
func (ctx *Context) Parse(val interface{}) {
	ctx.AbortIf(ctx.TryParse(val))
}

// TryParseParam decode val from Param
func (ctx *Context) TryParseParam(name string, val interface{}) error {
	return json.Unmarshal([]byte(ctx.Param(name)), val)
}

// ParseParam decode val from Param, if error != nil abort
func (ctx *Context) ParseParam(name string, val interface{}) {
	ctx.AbortIf(ctx.TryParseParam(name, val))
}

// TryParseQuery decode val from Query
func (ctx *Context) TryParseQuery(name string, val interface{}) error {
	return json.Unmarshal([]byte(ctx.Query(name)), val)
}

// ParseQuery decode val from Query, if error != nil abort
func (ctx *Context) ParseQuery(name string, val interface{}) {
	ctx.AbortIf(ctx.TryParseQuery(name, val))
}

// TryParseForm decode val from Form
func (ctx *Context) TryParseForm(name string, val interface{}) error {
	return json.Unmarshal([]byte(ctx.Form(name)), val)
}

// ParseForm decode val from Form, if error != nil abort
func (ctx *Context) ParseForm(name string, val interface{}) {
	ctx.AbortIf(ctx.TryParseForm(name, val))
}

// Abort WriteHeader 400 then abort
func (ctx *Context) Abort() {
	ctx.ResponseWriter.WriteHeader(defaultHTTPStatusError)
	panic(errors.New("Abort by user"))
}

// AbortIf if error != nill, WriteHeader 400 then abort
func (ctx *Context) AbortIf(err error) {
	if err != nil {
		ctx.ResponseWriter.WriteHeader(defaultHTTPStatusError)
		panic(err)
	}
}

// AbortFn if error != nill, call fn then abort
func (ctx *Context) AbortFn(code int, err error, fn func(code int, err error) error) {
	if err != nil {
		if fn != nil {
			fn(code, err)
		}
		panic(err)
	}
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
	ctx.ResponseWriter.WriteHeader(defaultHTTPStatusSuccess)
	return ctx.WriteJSON(data)
}

// WriteError with http 400 and code
func (ctx *Context) WriteError(code int, err error) error {
	data := &responseData{
		Success: false,
		Code:    code,
		Error:   err,
	}
	ctx.ResponseWriter.WriteHeader(defaultHTTPStatusError)
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
