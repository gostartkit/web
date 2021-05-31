package web

import "errors"

var (
	// ErrUnauthorized 401
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden 403
	ErrForbidden = errors.New("forbidden")
	// ErrUnsupportedType 415
	ErrUnsupportedType = errors.New("unsupported type")
)

// Param struct
type Param struct {
	Key   string
	Value string
}

// Params list
type Params []Param

// Val get value from Params by name
func (ps Params) Val(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}
