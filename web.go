package web

import "net/http"

type Handler func(*Context)

type Controller interface {
	Index(*Context)
	Create(*Context)
	Update(*Context)
	Delete(*Context)
}

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (params Params) val(name string) string {
	for i := range params {
		if params[i].Key == name {
			return params[i].Value
		}
	}
	return ""
}

type Context struct {
	Request *http.Request
	Params  *Params
	http.ResponseWriter
}

func (ctx *Context) Val(key string) string {
	return ctx.Params.val(key)
}

func (ctx *Context) WriteString(text string) {
	ctx.ResponseWriter.Write([]byte(text))
}
