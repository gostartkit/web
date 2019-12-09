package web

import "net/http"

// Controller interface
type Controller interface {
	Index(ctx *Context)
	Create(ctx *Context)
	Detail(ctx *Context)
	Update(ctx *Context)
	Destroy(ctx *Context)
}

// Validation interface
type Validation interface {
	Validate(r *http.Request) error
}

// ResponseData struct
type ResponseData struct {
	Success bool          `json:"success"`
	Code    int           `json:"code"`
	Result  interface{}   `json:"result"`
	Errors  ErrorMessages `json:"errors"`
}

// ErrorMessage struct
type ErrorMessage struct {
	Name   string  `json:"name"`
	Errors []error `json:"errors"`
}

// ErrorMessages Error Collection
type ErrorMessages []ErrorMessage
