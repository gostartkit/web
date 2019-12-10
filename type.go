package web

import (
	"fmt"
	"net/http"
	"strings"
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
	Validate(r *http.Request) ValidationError
}

// AttributeError struct
type AttributeError struct {
	Name  string `json:"name"`
	Error error  `json:"error"`
}

// ValidationError Error Collection
type ValidationError []AttributeError

// CreateValidationError return new ValidationError
func CreateValidationError() *ValidationError {
	return &ValidationError{}
}

func (validationError ValidationError) Error() string {
	var str strings.Builder

	for _, attributeError := range validationError {
		fmt.Fprintf(&str, "%s: %v", attributeError.Name, attributeError.Error)
	}

	return str.String()
}
