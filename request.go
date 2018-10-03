package web

// Request is type of a web.Request
type Request struct {
}

// Href is
func (r *Request) Href() {

}

// Path is
func (r *Request) Path() {

}

// QueryString is
func (r *Request) QueryString() {

}

// Search is
func (r *Request) Search() {

}

// Host is
func (r *Request) Host() {

}

// Hostname is
func (r *Request) Hostname() {

}

// URL is
func (r *Request) URL() {

}

// Type is
func (r *Request) Type() {

}

// Charset is
func (r *Request) Charset() {

}

// Query is
func (r *Request) Query() {

}

// Fresh is
func (r *Request) Fresh() {

}

// Stale is
func (r *Request) Stale() {

}

// Protocol is
func (r *Request) Protocol() {

}

// Secure is
func (r *Request) Secure() {

}

// IP is
func (r *Request) IP() {

}

// SubDomains is
func (r *Request) SubDomains() {

}

// IS is
func (r *Request) IS(types string) {

}

// Accepts is
func (r *Request) Accepts(types string) string {
	return ""
}

// AcceptsEncodings is
func (r *Request) AcceptsEncodings(encodings string) string {
	return ""
}

// AcceptsCharsets is
func (r *Request) AcceptsCharsets(charsets string) string {
	return ""
}

// AcceptsLanguages is
func (r *Request) AcceptsLanguages(langs string) string {
	return ""
}
