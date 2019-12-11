package web

const (
	defaultHTTPClientError int = 400
)

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

// ValidationError struct
type ValidationError struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func (r *ValidationError) Error() string {
	return r.Message
}

// Error struct
type Error struct {
	Message string `json:"message"`
}

func (r *Error) Error() string {
	return r.Message
}
