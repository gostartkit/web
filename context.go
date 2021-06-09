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
		w:          w,
		r:          r,
		params:     params,
		statusCode: 200,
	}

	return ctx
}

// Context is type of an web.Context
type Context struct {
	w           http.ResponseWriter
	r           *http.Request
	params      *Params
	urlValues   *url.Values
	userID      uint64
	accept      *string
	contentType *string
	statusCode  int
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
		urlValues := ctx.r.URL.Query()
		ctx.urlValues = &urlValues
	}

	return ctx.urlValues.Get(name)
}

// Form get value from Form
func (ctx *Context) Form(name string) string {
	if ctx.r.Form == nil {
		ctx.r.ParseForm()
	}
	return ctx.r.Form.Get(name)
}

// Path return ctx.r.URL.Path
func (ctx *Context) Path() string {
	return ctx.r.URL.Path
}

// Method return ctx.r.Method
func (ctx *Context) Method() string {
	return ctx.r.Method
}

// TryParseBody decode val from Request.Body
func (ctx *Context) TryParseBody(val interface{}) error {
	switch ctx.ContentType() {
	case "application/json":
		if err := json.NewDecoder(ctx.r.Body).Decode(val); err != nil {
			return err
		}
	case "application/x-gob":
		if err := gob.NewDecoder(ctx.r.Body).Decode(val); err != nil {
			return err
		}
	case "application/x-www-form-urlencoded":
		if err := formReader(ctx.r.Body, val); err != nil {
			return err
		}
	case "multipart/form-data":
		if err := formDataReader(ctx.r.Body, val); err != nil {
			return err
		}
	case "application/octet-stream":
		if err := binaryReader(ctx.r.Body, val); err != nil {
			return err
		}
	case "application/xml":
		if err := xml.NewDecoder(ctx.r.Body).Decode(val); err != nil {
			return err
		}
	default:
		return errors.New("tryParseBody(unsupported contentType '" + ctx.ContentType() + "')")
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

// writeBytes Write bytes
func (ctx *Context) writeBytes(val []byte) (int, error) {
	return ctx.w.Write(val)
}

// writeString Write String
func (ctx *Context) writeString(val string) (int, error) {
	return ctx.w.Write([]byte(val))
}

// write write data base on accept header
func (ctx *Context) write(val interface{}) error {
	switch ctx.Accept() {
	case "application/json":
		return ctx.writeJSON(val)
	case "application/x-gob":
		return ctx.writeGOB(val)
	case "application/xml":
		return ctx.writeXML(val)
	case "application/octet-stream":
		return ctx.writeBinary(val)
	default:
		if strings.HasPrefix(ctx.Accept(), "text/html") {
			return ctx.writeHTML(val)
		}
		return ctx.writeJSON(val)
	}
}

// writeJSON Write JSON
func (ctx *Context) writeJSON(val interface{}) error {
	return json.NewEncoder(ctx.w).Encode(val)
}

// writeXML Write XML
func (ctx *Context) writeXML(val interface{}) error {
	return xml.NewEncoder(ctx.w).Encode(val)
}

// writeGOB Write GOB
func (ctx *Context) writeGOB(val interface{}) error {
	return gob.NewEncoder(ctx.w).Encode(val)
}

// writeBinary Write Binary
func (ctx *Context) writeBinary(val interface{}) error {
	return binaryWriter(ctx.w, val)
}

// writeHTML Write HTML
func (ctx *Context) writeHTML(val interface{}) error {
	return htmlWriter(ctx.w, ctx, val)
}

// Status return status code
func (ctx *Context) Status() int {
	return ctx.statusCode
}

// SetStatus Write status code to header
func (ctx *Context) SetStatus(code int) {
	ctx.statusCode = code
	ctx.w.WriteHeader(code)
}

// Get get header, short hand for ctx.Request.Header.Get
func (ctx *Context) Get(key string) string {
	return ctx.r.Header.Get(key)
}

// Set set header, short hand for ctx.ResponseWriter.Header().Set
func (ctx *Context) Set(key string, value string) {
	ctx.w.Header().Set(key, value)
}

// Add add header, short hand for ctx.ResponseWriter.Header().Add
func (ctx *Context) Add(key string, value string) {
	ctx.w.Header().Add(key, value)
}

// Del del header, short hand for ctx.ResponseWriter.Header().Del
func (ctx *Context) Del(key string) {
	ctx.w.Header().Del(key)
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
	ctx.writeString("Redirecting to: " + url)
}
