package web

import "strings"

const (
	defaultHTTPSuccess int = 200
	defaultHTTPError   int = 400
)

// Middleware struct
type Middleware struct {
	Path     string
	Callback Callback
}

// Middlewares Middleware list
type Middlewares []Middleware

func (m Middlewares) exec(path string, ctx *Context) {
	for i := range m {
		if strings.HasPrefix(path, m[i].Path) {
			if m[i].Callback != nil {
				m[i].Callback(ctx)
			}
		}
	}
}

// ResponseData struct
type ResponseData struct {
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
