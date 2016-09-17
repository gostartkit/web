package web

import "net/http"

type Handle func(*Context)

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (params Params) Val(name string) string {
	for i := range params {
		if params[i].Key == name {
			return params[i].Value
		}
	}
	return ""
}

type Context struct {
	Request        *http.Request
	Params         *Params
	ResponseWriter http.ResponseWriter
}

func (ctx *Context) WriteString(content string) {
	ctx.ResponseWriter.Write([]byte(content))
}
