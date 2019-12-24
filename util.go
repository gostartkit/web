package web

import (
	"log"
	"mime"
	"net/http"
	"strings"
)

// newContext return a web.Context
func newContext(w http.ResponseWriter, r *http.Request, params *Params) *Context {

	ctx := &Context{
		ResponseWriter: w,
		Request:        r,
		params:         params,
	}

	return ctx
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

func logf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// cleanPath
// wrap path with / like: "user" return "/user/"
func cleanPath(path string) string {

	if len(path) > 0 {
		if path[0] != '/' {
			path = "/" + path
		}
	} else {
		path = "/"
	}

	pos := len(path) - 1

	if pos >= 0 {
		if path[pos] != '/' {
			path = path + "/"
		}
	} else {
		path = "/"
	}

	return path
}
