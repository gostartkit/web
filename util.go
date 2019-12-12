package web

import (
	"log"
	"mime"
	"net/http"
	"os"
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return !info.IsDir()
}
