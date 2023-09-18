package web

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// createCtx return a web.Ctx
func createCtx(w http.ResponseWriter, r *http.Request, params *Params) *Ctx {

	c := &Ctx{
		w:      w,
		r:      r,
		param:  params,
		parsed: false,
	}

	return c
}

// Ctx is type of an web.Ctx
type Ctx struct {
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
func (c *Ctx) Init(userID uint64, userRight int64) {
	c.userID = userID
	c.userRight = userRight
}

// UserID get userID
func (c *Ctx) UserID() uint64 {
	return c.userID
}

// UserRight get UserRight
func (c *Ctx) UserRight() int64 {
	return c.userRight
}

// Param get value from Params
func (c *Ctx) Param(name string) string {
	return c.param.Val(name)
}

// Query get value from QueryString
func (c *Ctx) Query(name string) string {
	if c.query == nil {
		query := c.r.URL.Query()
		c.query = &query
	}
	return c.query.Get(name)
}

// Form get value from Form
func (c *Ctx) Form(name string) string {
	if !c.parsed {
		c.r.ParseForm()
		c.parsed = true
	}
	return c.r.FormValue(name)
}

// Host return c.r.Host
func (c *Ctx) Host() string {
	return c.r.Host
}

// Path return c.r.URL.Path
func (c *Ctx) Path() string {
	return c.r.URL.Path
}

// Path return c.r.Body
func (c *Ctx) Body() io.ReadCloser {
	return c.r.Body
}

// Method return c.r.Method
func (c *Ctx) Method() string {
	return c.r.Method
}

// RemoteAddr return remote ip address
func (c *Ctx) RemoteAddr() string {
	return c.r.RemoteAddr
}

// UserAgent return User-Agent header
func (c *Ctx) UserAgent() string {
	return c.Get("User-Agent")
}

// IsAjax if X-Requested-With header is XMLHttpRequest return true, else false
func (c *Ctx) IsAjax() bool {
	return c.Get("X-Requested-With") == "XMLHttpRequest"
}

// TryParseBody decode val from Request.Body
func (c *Ctx) TryParseBody(val interface{}) error {
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
func (c *Ctx) TryParseParam(name string, val interface{}) error {
	return TryParse(c.Param(name), val)
}

// TryParseQuery decode val from Query
func (c *Ctx) TryParseQuery(name string, val interface{}) error {
	return TryParse(c.Query(name), val)
}

// TryParseForm decode val from Form
func (c *Ctx) TryParseForm(name string, val interface{}) error {
	return TryParse(c.Form(name), val)
}

// ParamInt decode val from Param by name
func (c *Ctx) ParamInt(name string) (int, error) {
	n, err := strconv.ParseInt(c.Param(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// ParamUint decode val from Param by name
func (c *Ctx) ParamUint(name string) (uint, error) {
	n, err := strconv.ParseUint(c.Param(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// ParamInt8 decode val from Param by name
func (c *Ctx) ParamInt8(name string) (int8, error) {
	n, err := strconv.ParseInt(c.Param(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(n), nil
}

// ParamUint8 decode val from Param by name
func (c *Ctx) ParamUint8(name string) (uint8, error) {
	n, err := strconv.ParseUint(c.Param(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(n), nil
}

// ParamInt16 decode val from Param by name
func (c *Ctx) ParamInt16(name string) (int16, error) {
	n, err := strconv.ParseInt(c.Param(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(n), nil
}

// ParamUint16 decode val from Param by name
func (c *Ctx) ParamUint16(name string) (uint16, error) {
	n, err := strconv.ParseUint(c.Param(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), nil
}

// ParamInt32 decode val from Param by name
func (c *Ctx) ParamInt32(name string) (int32, error) {
	n, err := strconv.ParseInt(c.Param(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}

// ParamUint32 decode val from Param by name
func (c *Ctx) ParamUint32(name string) (uint32, error) {
	n, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// ParamInt64 decode val from Param by name
func (c *Ctx) ParamInt64(name string) (int64, error) {
	n, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ParamUint64 decode val from Param by name
func (c *Ctx) ParamUint64(name string) (uint64, error) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ParamFloat32 decode val from Param by name
func (c *Ctx) ParamFloat32(name string) (float32, error) {
	n, err := strconv.ParseFloat(c.Param(name), 32)
	if err != nil {
		return 0, err
	}
	return float32(n), nil
}

// ParamFloat64 decode val from Param by name
func (c *Ctx) ParamFloat64(name string) (float64, error) {
	n, err := strconv.ParseFloat(c.Param(name), 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ParamBool decode val from Param by name
func (c *Ctx) ParamBool(name string) (bool, error) {
	n, err := strconv.ParseBool(c.Param(name))
	if err != nil {
		return false, err
	}
	return n, nil
}

// QueryInt decode val from Query by name
func (c *Ctx) QueryInt(name string) (int, error) {
	n, err := strconv.ParseInt(c.Query(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// QueryUint decode val from Query by name
func (c *Ctx) QueryUint(name string) (uint, error) {
	n, err := strconv.ParseUint(c.Query(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// QueryInt8 decode val from Query by name
func (c *Ctx) QueryInt8(name string) (int8, error) {
	n, err := strconv.ParseInt(c.Query(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(n), nil
}

// QueryUint8 decode val from Query by name
func (c *Ctx) QueryUint8(name string) (uint8, error) {
	n, err := strconv.ParseUint(c.Query(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(n), nil
}

// QueryInt16 decode val from Query by name
func (c *Ctx) QueryInt16(name string) (int16, error) {
	n, err := strconv.ParseInt(c.Query(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(n), nil
}

// QueryUint16 decode val from Query by name
func (c *Ctx) QueryUint16(name string) (uint16, error) {
	n, err := strconv.ParseUint(c.Query(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), nil
}

// QueryInt32 decode val from Query by name
func (c *Ctx) QueryInt32(name string) (int32, error) {
	n, err := strconv.ParseInt(c.Query(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}

// QueryUint32 decode val from Query by name
func (c *Ctx) QueryUint32(name string) (uint32, error) {
	n, err := strconv.ParseUint(c.Query(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// QueryInt64 decode val from Query by name
func (c *Ctx) QueryInt64(name string) (int64, error) {
	n, err := strconv.ParseInt(c.Query(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// QueryUint64 decode val from Query by name
func (c *Ctx) QueryUint64(name string) (uint64, error) {
	n, err := strconv.ParseUint(c.Query(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// QueryFloat32 decode val from Query by name
func (c *Ctx) QueryFloat32(name string) (float32, error) {
	n, err := strconv.ParseFloat(c.Query(name), 32)
	if err != nil {
		return 0, err
	}
	return float32(n), nil
}

// QueryFloat64 decode val from Query by name
func (c *Ctx) QueryFloat64(name string) (float64, error) {
	n, err := strconv.ParseFloat(c.Query(name), 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// QueryBool decode val from Query by name
func (c *Ctx) QueryBool(name string) (bool, error) {
	n, err := strconv.ParseBool(c.Query(name))
	if err != nil {
		return false, err
	}
	return n, nil
}

// FormIn decode val from Form by name
func (c *Ctx) FormInt(name string) (int, error) {
	n, err := strconv.ParseInt(c.Form(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// FormUint decode val from Form by name
func (c *Ctx) FormUint(name string) (uint, error) {
	n, err := strconv.ParseUint(c.Form(name), 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// FormInt8 decode val from Form by name
func (c *Ctx) FormInt8(name string) (int8, error) {
	n, err := strconv.ParseInt(c.Form(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(n), nil
}

// FormUint8 decode val from Form by name
func (c *Ctx) FormUint8(name string) (uint8, error) {
	n, err := strconv.ParseUint(c.Form(name), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(n), nil
}

// FormInt16 decode val from Form by name
func (c *Ctx) FormInt16(name string) (int16, error) {
	n, err := strconv.ParseInt(c.Form(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(n), nil
}

// FormUint16 decode val from Form by name
func (c *Ctx) FormUint16(name string) (uint16, error) {
	n, err := strconv.ParseUint(c.Form(name), 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), nil
}

// FormInt32 decode val from Form by name
func (c *Ctx) FormInt32(name string) (int32, error) {
	n, err := strconv.ParseInt(c.Form(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}

// FormUint32 decode val from Form by name
func (c *Ctx) FormUint32(name string) (uint32, error) {
	n, err := strconv.ParseUint(c.Form(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// FormInt64 decode val from Form by name
func (c *Ctx) FormInt64(name string) (int64, error) {
	n, err := strconv.ParseInt(c.Form(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// FormUint64 decode val from Form by name
func (c *Ctx) FormUint64(name string) (uint64, error) {
	n, err := strconv.ParseUint(c.Form(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// FormFloat32 decode val from Form by name
func (c *Ctx) FormFloat32(name string) (float32, error) {
	n, err := strconv.ParseFloat(c.Form(name), 32)
	if err != nil {
		return 0, err
	}
	return float32(n), nil
}

// FormFloat64 decode val from Form by name
func (c *Ctx) FormFloat64(name string) (float64, error) {
	n, err := strconv.ParseFloat(c.Form(name), 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// FormBool decode val from Form by name
func (c *Ctx) FormBool(name string) (bool, error) {
	n, err := strconv.ParseBool(c.Form(name))
	if err != nil {
		return false, err
	}
	return n, nil
}

// Write Write data base on accept header
func (c *Ctx) Write(val interface{}) error {

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
func (c *Ctx) WriteJSON(val interface{}) error {
	return json.NewEncoder(c.w).Encode(val)
}

// WriteXML Write XML
func (c *Ctx) WriteXML(val interface{}) error {
	return xml.NewEncoder(c.w).Encode(val)
}

// WriteGOB Write GOB
func (c *Ctx) WriteGOB(val interface{}) error {
	return gob.NewEncoder(c.w).Encode(val)
}

// WriteBinary Write Binary
func (c *Ctx) WriteBinary(val interface{}) error {
	return ErrMethodNotImplemented
}

// WriteAvro Write Avro
func (c *Ctx) WriteAvro(val interface{}) error {
	return ErrMethodNotImplemented
}

// SetLocation set Location with status code
func (c *Ctx) SetLocation(url string) {
	c.Set("Location", url)
}

// Get get header, short hand for c.Request.Header.Get
func (c *Ctx) Get(key string) string {
	return c.r.Header.Get(key)
}

// Set set header, short hand for c.ResponseWriter.Header().Set
func (c *Ctx) Set(key string, value string) {
	c.w.Header().Set(key, value)
}

// Add add header, short hand for c.ResponseWriter.Header().Add
func (c *Ctx) Add(key string, value string) {
	c.w.Header().Add(key, value)
}

// Del del header, short hand for c.ResponseWriter.Header().Del
func (c *Ctx) Del(key string) {
	c.w.Header().Del(key)
}

// Accept get Accept from header
func (c *Ctx) Accept() string {
	if c.accept == nil {
		ac := c.Get("Accept")
		c.accept = &ac
	}
	return *c.accept
}

// ContentType get Content-Type from header
func (c *Ctx) ContentType() string {
	if c.contentType == nil {
		ctype := c.Get("Content-Type")
		c.contentType = &ctype
	}
	return *c.contentType
}

// SetContentType Set Content-Type to header
func (c *Ctx) SetContentType(val string) {
	c.Set("Content-Type", contentType(val))
}

// AcceptContentType set 'Accept' header to 'Content-Type' header
func (c *Ctx) AcceptContentType() {
	ac := c.Accept()
	switch ac {
	case "application/json", "application/octet-stream", "application/x-avro", "application/x-gob", "application/xml":
		c.SetContentType(ac)
	}
}
