package web

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

const (
	pbkdf2Iterations = 64000
	keySize          = 32
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

// Get value from Params by key
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

func (ctx *Context) SetSecureCookie(name string, val string, age int64) error {
	// server := ctx.Server
	// if len(server.Config.CookieSecret) == 0 {
	// 	return ErrMissingCookieSecret
	// }
	// if len(server.encKey) == 0 || len(server.signKey) == 0 {
	// 	return ErrInvalidKey
	// }
	// ciphertext, err := encrypt([]byte(val), server.encKey)
	// if err != nil {
	// 	return err
	// }
	// sig := sign(ciphertext, server.signKey)
	// data := base64.StdEncoding.EncodeToString(ciphertext) + "|" + base64.StdEncoding.EncodeToString(sig)
	// ctx.SetCookie(newCookie(name, data, age))
	return nil
}

func (ctx *Context) GetSecureCookie(name string) (string, bool) {
	// for _, cookie := range ctx.Request.Cookies() {
	// 	if cookie.Name != name {
	// 		continue
	// 	}
	// 	parts := strings.SplitN(cookie.Value, "|", 2)
	// 	if len(parts) != 2 {
	// 		return "", false
	// 	}
	// 	ciphertext, err := base64.StdEncoding.DecodeString(parts[0])
	// 	if err != nil {
	// 		return "", false
	// 	}
	// 	sig, err := base64.StdEncoding.DecodeString(parts[1])
	// 	if err != nil {
	// 		return "", false
	// 	}
	// 	expectedSig := sign([]byte(ciphertext), ctx.Server.signKey)
	// 	if !bytes.Equal(expectedSig, sig) {
	// 		return "", false
	// 	}
	// 	plaintext, err := decrypt(ciphertext, ctx.Server.encKey)
	// 	if err != nil {
	// 		return "", false
	// 	}
	// 	return string(plaintext), true
	// }
	return "", false
}
