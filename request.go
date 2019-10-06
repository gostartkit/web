package web

import "net/http"

// Request is type of a web.Request
type Request struct {
	r *http.Request
}

// Header is type of Request.Header
func (o *Request) Header() {

}

// Headers is type of Request.Headers
func (o *Request) Headers() {

}

// Method is type of Request.Method
func (o *Request) Method() {

}

// Length Return Content-Length as a number when present, or -1
func (o *Request) Length() int {
	return 0
}

// URL Get request URL.
func (o *Request) URL() {

}

// Href is
func (o *Request) Href() {

}

// Path is
func (o *Request) Path() {

}

// QueryString is
func (o *Request) QueryString() {

}

// Search is
func (o *Request) Search() {

}

// Host is
func (o *Request) Host() {

}

// Hostname is
func (o *Request) Hostname() {

}

// Type is
func (o *Request) Type() {

}

// Charset is
func (o *Request) Charset() {

}

// Query is
func (o *Request) Query() {

}

// Fresh is
func (o *Request) Fresh() {

}

// Stale is
func (o *Request) Stale() {

}

// Protocol is
func (o *Request) Protocol() {

}

// Secure is
func (o *Request) Secure() {

}

// IP is
func (o *Request) IP() {

}

// SubDomains is
func (o *Request) SubDomains() {

}

// Is is
func (o *Request) Is(types string) {

}

// Accepts is
func (o *Request) Accepts(types string) string {
	return ""
}

// AcceptsEncodings is
func (o *Request) AcceptsEncodings(encodings string) string {
	return ""
}

// AcceptsCharsets is
func (o *Request) AcceptsCharsets(charsets string) string {
	return ""
}

// AcceptsLanguages is
func (o *Request) AcceptsLanguages(langs string) string {
	return ""
}
