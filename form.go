package web

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

const (
	maxFormSize      int = 10 << 20 // 10 MB is a lot of text.
	formBufSize      int = 512
	formKeyBufSize   int = 32
	formValueBufSize int = 64
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

// formReader decode data from form
// ContentType: application/x-www-form-urlencoded
func formReader(r io.ReadCloser, v interface{}) error {
	if _formReader != nil {
		return _formReader(r, v)
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

	rt := rv.Type()

	m := make(map[string]int)

	for i := 0; i < rt.NumField(); i++ {
		tag := rt.Field(i).Tag.Get("json")
		if len(tag) > 0 && tag != "-" {
			m[tag] = i
		}
	}

	// stream
	formSize := 0

	buf := make([]byte, 0, formBufSize)

	var (
		key = make([]byte, 0, formKeyBufSize)
		val = make([]byte, 0, formValueBufSize)
	)

	isKey := true

	for {
		prev := 0
		n, err := r.Read(buf[0:formBufSize])

		if err != nil {

			if err == io.EOF {
				err = nil
			}

			if err != nil {
				return err
			}
		}

		formSize += n

		if formSize > maxFormSize {
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

				if err := formKevValue(key, val, &m, &rv); err != nil {
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

		if n != formBufSize {
			break
		}
	}

	if len(key) > 0 {
		if err := formKevValue(key, val, &m, &rv); err != nil {
			return err
		}
	}

	return nil
}

func formKevValue(key []byte, value []byte, m *map[string]int, v *reflect.Value) error {

	k, err := queryUnescape(key)

	if err != nil {
		return err
	}

	if i, ok := (*m)[k]; ok {

		val, err := queryUnescape(value)

		if err != nil {
			return err
		}

		if len(val) > 0 {
			field := v.Field(i)

			if err := tryParse(val, &field); err != nil {
				return err
			}
		}
	}

	return nil
}

// formDataReader decode data from form
// ContentType: multipart/form-data
func formDataReader(r io.ReadCloser, v interface{}) error {
	if _formDataReader != nil {
		return _formDataReader(r, v)
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
