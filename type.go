package web

import (
	"fmt"
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
	Validate() ValidationError
}

// AttributeError struct
type AttributeError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

// ValidationError Error Collection
type ValidationError []AttributeError

func (validationError ValidationError) Error() string {
	var str strings.Builder

	for _, attributeError := range validationError {
		fmt.Fprintf(&str, "%s: %v", attributeError.Name, attributeError.Error)
	}

	return str.String()
}
