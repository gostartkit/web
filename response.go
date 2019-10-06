package web

import "net/http"

// Response is type of an web.Response
type Response struct {
	w http.ResponseWriter
}

func (o *Response) Write(val string) {
	o.w.Write([]byte(val))
}
