package web

import "errors"

var (
	// ErrUnauthorized 401
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden 403
	ErrForbidden = errors.New("forbidden")
)

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
