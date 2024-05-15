package web

import "net/http"

// IRelease
type IRelease interface {
	Release()
}

// Callback
type Callback func(c *Ctx) (any, error)

// CorsCallback
type CorsCallback func(set func(key string, value string), origin string, allow []string)

// PanicCallback
type PanicCallback func(http.ResponseWriter, *http.Request, any)

// Middleware
type Middleware func(Callback) Callback

// Chain middleware chain
type Chain []Middleware

// Reader
type Reader func(c *Ctx, v any) error

// Writer
type Writer func(c *Ctx, v any) error

// Param struct
type Param struct {
	Key   string
	Value string
}

// Params list
type Params []Param

// Val get value from Params by name
func (o *Params) Val(name string) string {
	for i := range *o {
		if (*o)[i].Key == name {
			return (*o)[i].Value
		}
	}
	return ""
}
