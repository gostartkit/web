package web

import "net/http"

// Any is an alias for interface{}
type Any interface{}

// IRelease
type IRelease interface {
	Release()
}

// Callback function
type Callback func(c *Ctx) (Any, error)

// CorsCallback function
type CorsCallback func(set func(key string, value string), origin *string, allow []string)

// PanicCallback function
type PanicCallback func(http.ResponseWriter, *http.Request, Any)

// Middleware
type Middleware func(Callback) Callback

// Chain middleware chain
type Chain []Middleware

// Reader function
type Reader func(c *Ctx, v Any) error

// Writer function
type Writer func(c *Ctx, v Any) error

// Param struct
type Param struct {
	Key   string
	Value string
}

// Params list
type Params []Param

// Val get value from Params by name
func (o Params) Val(name string) string {
	for i := range o {
		if o[i].Key == name {
			return o[i].Value
		}
	}
	return ""
}
