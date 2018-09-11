package web

// Context is type of an web Context
type Context struct {
	Application *Application
	Response    *Response
	Request     *Request
}

func (ctx *Context) Cookies() {

}
