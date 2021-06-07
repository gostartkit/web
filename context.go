package web

import (
	"encoding/gob"
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
	userID         uint64
	accept         *string
	contentType    *string
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
	switch ctx.ContentType() {
	case "application/json":
		if err := json.NewDecoder(ctx.Request.Body).Decode(val); err != nil {
			return err
		}
	case "application/x-gob":
		if err := gob.NewDecoder(ctx.Request.Body).Decode(val); err != nil {
			return err
		}
	case "application/x-www-form-urlencoded":
		if err := formReader(ctx.Request.Body, val); err != nil {
			return err
		}
	case "multipart/form-data":
		if err := formDataReader(ctx.Request.Body, val); err != nil {
			return err
		}
	case "application/octet-stream":
		if err := binaryReader(ctx.Request.Body, val); err != nil {
			return err
		}
	case "application/xml":
		if err := xml.NewDecoder(ctx.Request.Body).Decode(val); err != nil {
			return err
		}
	}
	return nil
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

// WriteBytes Write bytes
func (ctx *Context) WriteBytes(val []byte) (int, error) {
	return ctx.ResponseWriter.Write(val)
}

// WriteString Write String
func (ctx *Context) WriteString(val string) (int, error) {
	return ctx.ResponseWriter.Write([]byte(val))
}

// Write Write data base on accept header
func (ctx *Context) Write(val interface{}) error {
	switch ctx.Accept() {
	case "application/json":
		return ctx.WriteJSON(val)
	case "application/x-gob":
		return ctx.WriteGOB(val)
	case "application/xml":
		return ctx.WriteXML(val)
	case "application/octet-stream":
		return ctx.WriteBinary(val)
	default:
		return ctx.WriteJSON(val)
	}
}

// WriteJSON Write JSON
func (ctx *Context) WriteJSON(val interface{}) error {
	return json.NewEncoder(ctx.ResponseWriter).Encode(val)
}

// WriteXML Write XML
func (ctx *Context) WriteXML(val interface{}) error {
	return xml.NewEncoder(ctx.ResponseWriter).Encode(val)
}

// WriteGOB Write GOB
func (ctx *Context) WriteGOB(val interface{}) error {
	return gob.NewEncoder(ctx.ResponseWriter).Encode(val)
}

// WriteBinary Write Binary
func (ctx *Context) WriteBinary(val interface{}) error {
	return binaryWriter(ctx.ResponseWriter, val)
}

// Status Write status code to header
func (ctx *Context) Status(code int) {
	ctx.ResponseWriter.WriteHeader(code)
}

// Get get header, short hand for ctx.Request.Header.Get
func (ctx *Context) Get(key string) string {
	return ctx.Request.Header.Get(key)
}

// Set set header, short hand for ctx.ResponseWriter.Header().Set
func (ctx *Context) Set(key string, value string) {
	ctx.ResponseWriter.Header().Set(key, value)
}

// Add add header, short hand for ctx.ResponseWriter.Header().Add
func (ctx *Context) Add(key string, value string) {
	ctx.ResponseWriter.Header().Add(key, value)
}

// Del del header, short hand for ctx.ResponseWriter.Header().Del
func (ctx *Context) Del(key string) {
	ctx.ResponseWriter.Header().Del(key)
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
	ctx.ResponseWriter.WriteHeader(code)
	ctx.WriteString("Redirecting to: " + url)
}
