package web

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"
)

const (
	_maxFormSize      int = 10 << 20 // 10 MB is a lot of text.
	_formBufSize      int = 512
	_formKeyBufSize   int = 32
	_formValueBufSize int = 64
)

// formReader decode data from request body
// ContentType: application/x-www-form-urlencoded
func formReader(ctx *Context, v Data) error {
	if app().formReader != nil {
		return app().formReader(ctx, v)
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

		if ctx.form, err = parseForm(ctx.R.Body); err != nil {
			return err
		}
	}

	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {

		tagName := rt.Field(i).Tag.Get("web")

		var val string

		if len(tagName) > 0 {
			if tagName != "-" {
				val = ctx.form.Get(tagName)
			}
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
	if app().formDataReader != nil {
		return app().formDataReader(ctx, v)
	}
	return ErrFormDataReaderNotImplemented
}

// parseForm parse form from ctx.r.Body
func parseForm(r io.ReadCloser) (*url.Values, error) {
	m := make(url.Values)
	err := parseQuery(r, func(key, value []byte) error {
		k, err := queryUnescape(key)

		if err != nil {
			return err
		}

		val, err := queryUnescape(value)

		if err != nil {
			return err
		}

		m[k] = append(m[k], val)

		return nil
	})
	return &m, err
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

// parseQuery parse form from stream
// callback fn when got key, value
// application/x-www-form-urlencoded
func parseQuery(r io.ReadCloser, fn func(key []byte, value []byte) error) error {
	formSize := 0

	buf := make([]byte, 0, _formBufSize)

	var (
		key []byte = make([]byte, 0, _formKeyBufSize)
		val []byte = make([]byte, 0, _formValueBufSize)
	)

	isKey := true

	for {
		prev := 0
		n, err := r.Read(buf[0:_formBufSize])

		if err != nil {

			if err == io.EOF {
				err = nil
			}

			if err != nil {
				return err
			}
		}

		formSize += n

		if formSize > _maxFormSize {
			return errors.New("http: POST too large")
		}

		buf = buf[:n]

		for i := 0; i < n; i++ {
			r := buf[i]
			switch r {
			case '&', ';':
				if i > prev {
					val = append(val, buf[prev:i]...)
				}

				if err := fn(key, val); err != nil {
					return err
				}

				key = key[0:0]
				val = val[0:0]
				prev = i + 1
				isKey = true
			case '=':
				if i > prev {
					key = append(key, buf[prev:i]...)
				}
				prev = i + 1
				isKey = false
			}
		}

		if prev < n {
			if isKey {
				key = append(key, buf[prev:]...)
			} else {
				val = append(val, buf[prev:]...)
			}
		}

		if n != _formBufSize {
			break
		}
	}

	if len(key) > 0 {
		if err := fn(key, val); err != nil {
			return err
		}
	}

	return nil
}
