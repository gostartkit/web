package web

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
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

// TryParse try parse val to v
func (ctx *Context) TryParse(val string, v interface{}) error {
	if v == nil {
		return errors.New("TryParse(nil)")
	}

	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return errors.New("TryParse(non-pointer " + reflect.TypeOf(v).String() + ")")
	}

	if rv.IsNil() {
		return errors.New("TryParse(nil)")
	}

	for rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	if !rv.CanSet() {
		return errors.New("TryParse(can not set value to v)")
	}

	switch rv.Interface().(type) {
	case string:
		rv.SetString(val)
		return nil
	case int, int64:
		d, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		rv.SetInt(d)
		return nil
	case int32:
		d, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return err
		}
		rv.SetInt(d)
		return nil
	default:
		return json.Unmarshal([]byte(val), v)
	}
}

// Parse parse val to v, if error abort
func (ctx *Context) Parse(val string, v interface{}) {
	ctx.Abort(ctx.TryParse(val, v))
}

// TryParseBody decode val from Request.Body
func (ctx *Context) TryParseBody(val interface{}) error {
	if err := json.NewDecoder(ctx.Request.Body).Decode(val); err != nil {
		return err
	}
	defer ctx.Request.Body.Close()
	return nil
}

// ParseBody decode val from Request.Body, if error abort
func (ctx *Context) ParseBody(val interface{}) {
	ctx.Abort(ctx.TryParseBody(val))
}

// TryParseParam decode val from Query
func (ctx *Context) TryParseParam(name string, val interface{}) error {
	return ctx.TryParse(ctx.Param(name), val)
}

// ParseParam decode val from Param, if error abort
func (ctx *Context) ParseParam(name string, val interface{}) {
	ctx.Abort(ctx.TryParseParam(name, val))
}

// TryParseQuery decode val from Query
func (ctx *Context) TryParseQuery(name string, val interface{}) error {
	return ctx.TryParse(ctx.Query(name), val)
}

// ParseQuery decode val from Query, if error abort
func (ctx *Context) ParseQuery(name string, val interface{}) {
	ctx.Abort(ctx.TryParseQuery(name, val))
}

// TryParseForm decode val from Form
func (ctx *Context) TryParseForm(name string, val interface{}) error {
	return ctx.TryParse(ctx.Form(name), val)
}

// ParseForm decode val from Form, if error abort
func (ctx *Context) ParseForm(name string, val interface{}) {
	ctx.Abort(ctx.TryParseForm(name, val))
}

// Abort if error response err message with status 400 then abort
func (ctx *Context) Abort(err error) {
	if err != nil {
		ctx.WriteHeader(defaultHTTPError)
		ctx.WriteString(err.Error())
		panic(err)
	}
}

// AbortIf if error response err message with status 400 then abort
// else response val
func (ctx *Context) AbortIf(val interface{}, err error) {
	if err != nil {
		ctx.WriteHeader(defaultHTTPError)
		ctx.WriteString(err.Error())
		panic(err)
	} else {
		ctx.WriteJSON(val)
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
	ctx.SetHeader("Content-Type", contentType(val))
}

// Redirect to url with status
func (ctx *Context) Redirect(status int, url string) {
	ctx.SetHeader("Location", url)
	ctx.WriteHeader(status)
	ctx.WriteString("Redirecting to: " + url)
}
