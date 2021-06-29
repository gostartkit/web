package web

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	_maxFormSize      int = 10 << 20 // 10 MB is a lot of text.
	_formBufSize      int = 512
	_formKeyBufSize   int = 32
	_formValueBufSize int = 64
)

var (
	_formReader     Reader
	_formDataReader Reader
)

// SetFormReader set formReader
func SetFormReader(r Reader) {
	_formReader = r
}

// SetFormDataReader set formDataReader
func SetFormDataReader(r Reader) {
	_formDataReader = r
}

// formReader decode data from request body
// ContentType: application/x-www-form-urlencoded
func formReader(ctx *Context, v Data) error {
	if _formReader != nil {
		return _formReader(ctx, v)
	}

	if v == nil {
		return errors.New("formReader(nil)")
	}

	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return errors.New("formReader(non-pointer " + rv.Type().String() + ")")
	}

	if rv.IsNil() {
		return errors.New("formReader(nil)")
	}

	for rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("formReader(unsupported type '%s')", rv.Type().String())
	}

	if ctx.form == nil {

		var err error

		if ctx.form, err = ctx.parseForm(); err != nil {
			return err
		}
	}

	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {

		tagName := rt.Field(i).Tag.Get("web")

		var val string

		if len(tagName) > 0 {
			val = ctx.form.Get(tagName)
		} else {
			val = ctx.form.Get(rt.Field(i).Name)
		}

		if val != "" {
			field := rv.Field(i)
			if err := tryParse(val, &field); err != nil {
				return err
			}
		}
	}

	return nil
}

// formDataReader decode data from form
// ContentType: multipart/form-data
func formDataReader(ctx *Context, v Data) error {
	if _formDataReader != nil {
		return _formDataReader(ctx, v)
	}
	return errors.New("formDataReader not implemented")
}

// queryUnescape unescapes a string;
func queryUnescape(s []byte) (string, error) {
	n := 0
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			n++
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				s = s[i:]
				if len(s) > 3 {
					s = s[:3]
				}
				return "", fmt.Errorf("invalid query escape %s", s)
			}
			i += 3
		case '+':
			s[i] = ' '
			i++
		default:
			i++
		}
	}

	if n == 0 {
		return string(s), nil
	}

	var sb strings.Builder
	sb.Grow(len(s) - 2*n)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '%':
			sb.WriteByte(unhex(s[i+1])<<4 | unhex(s[i+2]))
			i += 2
		case '+':
			sb.WriteByte(' ')
		default:
			sb.WriteByte(s[i])
		}
	}
	return sb.String(), nil
}
