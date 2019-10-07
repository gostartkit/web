package web

// Context is type of an web.Context
type Context struct {
	Response *Response
	Request  *Request
	Params   *Params
}

// Cookies is
func (ctx *Context) Cookies() {

}

// Throw is
func (ctx *Context) Throw() {

}

// Assert is
func (ctx *Context) Assert() {

}
