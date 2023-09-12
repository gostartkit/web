package web

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// createWebContext return a web.Context
func createWebContext(w http.ResponseWriter, r *http.Request, params *Params) *WebContext {

	c := &WebContext{
		w:      w,
		r:      r,
		param:  params,
		parsed: false,
	}

	return c
}

// WebContext is type of an web.WebContext
type WebContext struct {
	w           http.ResponseWriter
	r           *http.Request
	param       *Params
	query       *url.Values
	parsed      bool
	userID      uint64
	userRight   int64
	accept      *string
	contentType *string
}

// Init init context
func (c *WebContext) Init(userID uint64, userRight int64) {
	c.userID = userID
	c.userRight = userRight
}

// UserID get userID
func (c *WebContext) UserID() uint64 {
	return c.userID
}

// UserRight get UserRight
func (c *WebContext) UserRight() int64 {
	return c.userRight
}

// Param get value from Params
func (c *WebContext) Param(name string) string {
	return c.param.Val(name)
}

// Query get value from QueryString
func (c *WebContext) Query(name string) string {
	if c.query == nil {
		query := c.r.URL.Query()
		c.query = &query
	}
	return c.query.Get(name)
}

// Form get value from Form
func (c *WebContext) Form(name string) string {
	if !c.parsed {
		c.r.ParseForm()
		c.parsed = true
	}
	return c.r.FormValue(name)
}

// Host return c.r.Host
func (c *WebContext) Host() string {
	return c.r.Host
}

// Path return c.r.URL.Path
func (c *WebContext) Path() string {
	return c.r.URL.Path
}

// Path return c.r.Body
func (c *WebContext) Body() io.ReadCloser {
	return c.r.Body
}

// Method return c.r.Method
func (c *WebContext) Method() string {
	return c.r.Method
}

// RemoteAddr return remote ip address
func (c *WebContext) RemoteAddr() string {
	return c.r.RemoteAddr
}

// UserAgent return User-Agent header
func (c *WebContext) UserAgent() string {
	return c.Get("User-Agent")
}

// IsAjax if X-Requested-With header is XMLHttpRequest return true, else false
func (c *WebContext) IsAjax() bool {
	return c.Get("X-Requested-With") == "XMLHttpRequest"
}

// TryParseBody decode val from Request.Body
func (c *WebContext) TryParseBody(val interface{}) error {
	switch {
	case strings.HasPrefix(c.ContentType(), "application/json"):
		return json.NewDecoder(c.r.Body).Decode(val)
	case strings.HasPrefix(c.ContentType(), "application/x-gob"):
		return gob.NewDecoder(c.r.Body).Decode(val)
	case strings.HasPrefix(c.ContentType(), "application/octet-stream"):
		return ErrContentTypeNotSupported
	case strings.HasPrefix(c.ContentType(), "application/xml"):
		return xml.NewDecoder(c.r.Body).Decode(val)
	default:
		return ErrContentTypeNotSupported
	}
}

// TryParseParam decode val from Query
func (c *WebContext) TryParseParam(name string, val interface{}) error {
	return TryParse(c.Param(name), val)
}

// TryParseQuery decode val from Query
func (c *WebContext) TryParseQuery(name string, val interface{}) error {
	return TryParse(c.Query(name), val)
}

// TryParseForm decode val from Form
func (c *WebContext) TryParseForm(name string, val interface{}) error {
	return TryParse(c.Form(name), val)
}

// Write Write data base on accept header
func (c *WebContext) Write(val interface{}) error {

	switch c.Accept() {
	case "application/json":
		return c.WriteJSON(val)
	case "application/x-gob":
		return c.WriteGOB(val)
	case "application/octet-stream":
		return c.WriteBinary(val)
	case "application/x-avro":
		return c.WriteAvro(val)
	case "application/xml":
		return c.WriteXML(val)
	default:
		return c.WriteJSON(val)
	}
}

// WriteJSON Write JSON
func (c *WebContext) WriteJSON(val interface{}) error {
	return json.NewEncoder(c.w).Encode(val)
}

// WriteXML Write XML
func (c *WebContext) WriteXML(val interface{}) error {
	return xml.NewEncoder(c.w).Encode(val)
}

// WriteGOB Write GOB
func (c *WebContext) WriteGOB(val interface{}) error {
	return gob.NewEncoder(c.w).Encode(val)
}

// WriteBinary Write Binary
func (c *WebContext) WriteBinary(val interface{}) error {
	return ErrMethodNotImplemented
}

// WriteAvro Write Avro
func (c *WebContext) WriteAvro(val interface{}) error {
	return ErrMethodNotImplemented
}

// SetLocation set Location with status code
func (c *WebContext) SetLocation(url string) {
	c.Set("Location", url)
}

// Get get header, short hand for c.Request.Header.Get
func (c *WebContext) Get(key string) string {
	return c.r.Header.Get(key)
}

// Set set header, short hand for c.ResponseWriter.Header().Set
func (c *WebContext) Set(key string, value string) {
	c.w.Header().Set(key, value)
}

// Add add header, short hand for c.ResponseWriter.Header().Add
func (c *WebContext) Add(key string, value string) {
	c.w.Header().Add(key, value)
}

// Del del header, short hand for c.ResponseWriter.Header().Del
func (c *WebContext) Del(key string) {
	c.w.Header().Del(key)
}

// Accept get Accept from header
func (c *WebContext) Accept() string {
	if c.accept == nil {
		ac := c.Get("Accept")
		c.accept = &ac
	}
	return *c.accept
}

// ContentType get Content-Type from header
func (c *WebContext) ContentType() string {
	if c.contentType == nil {
		ctype := c.Get("Content-Type")
		c.contentType = &ctype
	}
	return *c.contentType
}

// SetContentType Set Content-Type to header
func (c *WebContext) SetContentType(val string) {
	c.Set("Content-Type", contentType(val))
}

// AcceptContentType set 'Accept' header to 'Content-Type' header
func (c *WebContext) AcceptContentType() {
	ac := c.Accept()
	switch ac {
	case "application/json", "application/octet-stream", "application/x-avro", "application/x-gob", "application/xml":
		c.SetContentType(ac)
	}
}
