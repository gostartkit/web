package web

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
	Validate() error
}

// responseData struct
type responseData struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Result  interface{} `json:"result"`
	Error   error       `json:"error"`
}

// ResponseError struct
type ResponseError struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

func (r *ResponseError) Error() string {
	return r.Text
}
