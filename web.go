package web

import (
	"encoding/json"
	"encoding/xml"
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

// Get value from Params
func (ctx *Context) Val(key string) string {
	return ctx.Params.val(key)
}

func (ctx *Context) WriteString(text string) (int, error) {
	return ctx.ResponseWriter.Write([]byte(text))
}

func (ctx *Context) WriteJson(v interface{}) (int, error) {
	b, err := json.Marshal(v)

	if err != nil {
		return 0, err
	}

	return ctx.ResponseWriter.Write(b)
}

func (ctx *Context) WriteXml(v interface{}) (int, error) {
	b, err := xml.Marshal(v)

	if err != nil {
		return 0, err
	}

	return ctx.ResponseWriter.Write(b)
}

func (ctx *Context) SetHeader(key string, value string, unique bool) {
	if unique {
		ctx.Header().Set(key, value)
	} else {
		ctx.Header().Add(key, value)
	}
}

func (ctx *Context) SetContentType(val string) {
	ctx.Header().Set("Content-Type", contentType(val))
}

func (ctx *Context) SetCookie(cookie *http.Cookie) {
	ctx.SetHeader("Set-Cookie", cookie.String(), false)
}

func contentType(val string) string {
	var ctype string
	if strings.ContainsRune(val, '/') {
		ctype = val
	} else {
		if !strings.HasPrefix(val, ".") {
			val = "." + val
		}
		ctype = mime.TypeByExtension(val)
	}
	return ctype
}
