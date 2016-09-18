package web

import (
	"mime"
	"net/http"
	"strings"
)

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

func (ctx *Context) ContentType(val string) string {
	var ctype string
	if strings.ContainsRune(val, '/') {
		ctype = val
	} else {
		if !strings.HasPrefix(val, ".") {
			val = "." + val
		}
		ctype = mime.TypeByExtension(val)
	}
	if ctype != "" {
		ctx.Header().Set("Content-Type", ctype)
	}
	return ctype
}

func (ctx *Context) SetHeader(key string, value string, unique bool) {
	if unique {
		ctx.Header().Set(key, value)
	} else {
		ctx.Header().Add(key, value)
	}
}

func (ctx *Context) SetCookie(cookie *http.Cookie) {
	ctx.SetHeader("Set-Cookie", cookie.String(), false)
}
