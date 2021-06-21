package web

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// createContext return a web.Context
func createContext(w http.ResponseWriter, r *http.Request, params *Params, query *url.Values) *Context {

	ctx := &Context{
		w:     w,
		r:     r,
		param: params,
		query: query,
		code:  200,
	}

	return ctx
}

// Context is type of an web.Context
type Context struct {
	w           http.ResponseWriter
	r           *http.Request
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
	return ctx.query.Get(name)
}

// Form get value from Form
func (ctx *Context) Form(name string) string {
	if ctx.form == nil {
		ctx.form, _ = ctx.parseForm()
	}
	return ctx.form.Get(name)
}

// Host return ctx.r.Host
func (ctx *Context) Host() string {
	return ctx.r.Host
}

// Path return ctx.r.URL.Path
func (ctx *Context) Path() string {
	return ctx.r.URL.Path
}

// Method return ctx.r.Method
func (ctx *Context) Method() string {
	return ctx.r.Method
}

// RemoteAddr return remote ip address
func (ctx *Context) RemoteAddr() string {
	return ctx.r.RemoteAddr
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
		return json.NewDecoder(ctx.r.Body).Decode(val)
	case strings.HasPrefix(ctx.ContentType(), "application/x-gob"):
		return gob.NewDecoder(ctx.r.Body).Decode(val)
	case strings.HasPrefix(ctx.ContentType(), "application/x-www-form-urlencoded"):
		return formReader(ctx, val)
	case strings.HasPrefix(ctx.ContentType(), "multipart/form-data"):
		return formDataReader(ctx, val)
	case strings.HasPrefix(ctx.ContentType(), "application/octet-stream"):
		return binaryReader(ctx, val)
	case strings.HasPrefix(ctx.ContentType(), "application/xml"):
		return xml.NewDecoder(ctx.r.Body).Decode(val)
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

// // writeBytes Write bytes
// func (ctx *Context) writeBytes(val []byte) (int, error) {
// 	return ctx.w.Write(val)
// }

// // writeString Write String
// func (ctx *Context) writeString(val string) (int, error) {
// 	return ctx.w.Write([]byte(val))
// }

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
	return ctx.code
}

// SetStatus Write status code to header
func (ctx *Context) SetStatus(code int) {
	ctx.code = code
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
}

// parseForm parse form from ctx.r.Body
func (ctx *Context) parseForm() (*url.Values, error) {
	m := make(url.Values)
	err := ctx.parseQuery(ctx.r.Body, func(key, value []byte) error {
		k, err := queryUnescape(key)

		if err != nil {
			return err
		}

		val, err := queryUnescape(value)

		if err != nil {
			return err
		}

		m[k] = append(m[k], val)

		return nil
	})
	return &m, err
}

// parseQuery parse form from stream
// callback fn when got key, value
// application/x-www-form-urlencoded
func (ctx *Context) parseQuery(r io.ReadCloser, fn func(key []byte, value []byte) error) error {
	formSize := 0

	buf := make([]byte, 0, _formBufSize)

	var (
		key []byte = make([]byte, 0, _formKeyBufSize)
		val []byte = make([]byte, 0, _formValueBufSize)
	)

	isKey := true

	for {
		prev := 0
		n, err := r.Read(buf[0:_formBufSize])

		if err != nil {

			if err == io.EOF {
				err = nil
			}

			if err != nil {
				return err
			}
		}

		formSize += n

		if formSize > _maxFormSize {
			return errors.New("http: POST too large")
		}

		buf = buf[:n]

		for i := 0; i < n; i++ {
			r := buf[i]
			switch r {
			case '&', ';':
				if i > prev {
					val = append(val, buf[prev:i]...)
				}

				if err := fn(key, val); err != nil {
					return err
				}

				key = key[0:0]
				val = val[0:0]
				prev = i + 1
				isKey = true
			case '=':
				if i > prev {
					key = append(key, buf[prev:i]...)
				}
				prev = i + 1
				isKey = false
			}
		}

		if prev < n {
			if isKey {
				key = append(key, buf[prev:]...)
			} else {
				val = append(val, buf[prev:]...)
			}
		}

		if n != _formBufSize {
			break
		}
	}

	if len(key) > 0 {
		if err := fn(key, val); err != nil {
			return err
		}
	}

	return nil
}
