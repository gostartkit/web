package web

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

const (
	pbkdf2Iterations = 64000
	keySize          = 32
)

type Handler func(*Context)

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (params Params) Get(name string) string {
	for i := range params {
		if params[i].Key == name {
			return params[i].Value
		}
	}
	return ""
}

type Context struct {
	Params  *Params
	Query   url.Values
	Payload url.Values
	Server  *Server
	Request *http.Request
	http.ResponseWriter
}

// Get value from Params by key
func (ctx *Context) Val(key string) string {
	return ctx.Params.Get(key)
}

// Get value from url Query by key
func (ctx *Context) Get(key string) string {

	if ctx.Query == nil {
		ctx.Query = ctx.Request.URL.Query()
	}

	return ctx.Query.Get(key)
}

// Get value from post Form by key
func (ctx *Context) Post(key string) string {

	if ctx.Payload == nil {
		ctx.Request.ParseForm()
		ctx.Payload = ctx.Request.PostForm
	}

	return ctx.Payload.Get(key)
}

func (ctx *Context) WriteString(text string) (int, error) {
	return ctx.ResponseWriter.Write([]byte(text))
}

func (ctx *Context) WriteJSON(v interface{}) error {
	return json.NewEncoder(ctx.ResponseWriter).Encode(v)
}

func (ctx *Context) WriteXML(v interface{}) error {
	return xml.NewEncoder(ctx.ResponseWriter).Encode(v)
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

func (ctx *Context) setCookie(cookie *http.Cookie) {
	ctx.SetHeader("Set-Cookie", cookie.String(), false)
}

func (ctx *Context) SetCookie(name string, val string, age int64) error {
	server := ctx.Server
	if len(server.cookieSecret) == 0 {
		return errors.New("cookieSecret empty")
	}
	if len(server.encKey) == 0 || len(server.signKey) == 0 {
		return errors.New("encKey or signKey empty")
	}
	ciphertext, err := encrypt([]byte(val), server.encKey)
	if err != nil {
		return err
	}
	sig := sign(ciphertext, server.signKey)
	data := base64.StdEncoding.EncodeToString(ciphertext) + "|" + base64.StdEncoding.EncodeToString(sig)
	ctx.setCookie(newCookie(name, data, age))
	return nil
}

func (ctx *Context) GetCookie(name string) string {
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name != name {
			continue
		}
		parts := strings.SplitN(cookie.Value, "|", 2)
		if len(parts) != 2 {
			return ""
		}
		ciphertext, err := base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			return ""
		}
		sig, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return ""
		}
		expectedSig := sign([]byte(ciphertext), ctx.Server.signKey)
		if !bytes.Equal(expectedSig, sig) {
			return ""
		}
		plaintext, err := decrypt(ciphertext, ctx.Server.encKey)
		if err != nil {
			return ""
		}
		return string(plaintext)
	}
	return ""
}

func (ctx *Context) CreateDatabase() (*sql.DB, error) {
	return sql.Open(ctx.Server.driverName, ctx.Server.dataSourceName)
}

type Controller interface {
	Init()
}
