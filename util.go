package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Reader function
type Reader func(r io.ReadCloser, v interface{}) error

// Writer function
type Writer func(io.Writer, interface{}) error

// HtmlWriter function
type HtmlWriter func(string, io.Writer, interface{}) error

const (
	maxFormSize int = 10 << 20 // 10 MB is a lot of text.
	formBufSize int = 512
)

var (
	_binaryReader   Reader
	_formReader     Reader
	_formDataReader Reader
	_binaryWriter   Writer
	_htmlWriter     HtmlWriter
)

// SetBinaryReader set binaryReader
func SetBinaryReader(r Reader) {
	_binaryReader = r
}

// SetFormReader set formReader
func SetFormReader(r Reader) {
	_formReader = r
}

// SetFormDataReader set formDataReader
func SetFormDataReader(r Reader) {
	_formDataReader = r
}

// SetBinaryWriter set binaryWriter
func SetBinaryWriter(w Writer) {
	_binaryWriter = w
}

// SetHtmlWriter set htmlWriter
func SetHtmlWriter(w HtmlWriter) {
	_htmlWriter = w
}

// TryParse try parse val to v
func TryParse(val string, v interface{}) error {

	if len(val) > 0 && (val[0] == '{' || val[0] == '[') {
		return json.Unmarshal([]byte(val), v)
	}

	if v == nil {
		return errors.New("TryParse(nil)")
	}

	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return errors.New("TryParse(non-pointer " + reflect.TypeOf(v).String() + ")")
	}

	if rv.IsNil() {
		return errors.New("TryParse(nil)")
	}

	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	if !rv.CanSet() {
		return errors.New("TryParse(can not set value to v)")
	}

	return tryParse(val, &rv)
}

// binaryReader decode data from binary
func binaryReader(r io.ReadCloser, v interface{}) error {
	if _binaryReader != nil {
		return _binaryReader(r, v)
	}
	return errors.New("binaryReader not implemented")
}

// binaryWriter encode data to binary
func binaryWriter(w io.Writer, v interface{}) error {
	if _binaryWriter != nil {
		return _binaryWriter(w, v)
	}
	return errors.New("binaryWriter not implemented")
}

// htmlWriter encode data to html
func htmlWriter(path string, w io.Writer, v interface{}) error {
	if _htmlWriter != nil {
		return _htmlWriter(path, w, v)
	}
	return errors.New("htmlWriter not implemented")
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
		key string
		val string
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
					val += string(buf[prev:i])
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
					key += string(buf[prev:i])
				}
				prev = i + 1
				isKey = false
			}
		}

		if prev < n {
			if isKey {
				key += string(buf[prev:])
			} else {
				val += string(buf[prev:])
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

func formKevValue(key string, value string, m *map[string]int, v *reflect.Value) error {
	var err error

	key, err = url.QueryUnescape(key)

	if err != nil {
		return err
	}

	if i, ok := (*m)[key]; ok {

		value, err = url.QueryUnescape(value)

		if err != nil {
			return err
		}

		if len(value) > 0 {
			field := v.Field(i)

			if err := tryParse(value, &field); err != nil {
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

// tryParse try parse val to v
func tryParse(val string, v *reflect.Value) error {

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		*v = v.Elem()
	}

	if !v.IsValid() {
		return errors.New("tryParse(rv invalid)")
	}

	if !v.CanSet() {
		return errors.New("tryParse(can not set value to rv)")
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(val)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		if v.OverflowInt(n) {
			return errors.New("tryParse(reflect.Value.OverflowInt)")
		}
		v.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		if v.OverflowUint(n) {
			return errors.New("tryParse(reflect.Value.OverflowUint)")
		}
		v.SetUint(n)
		return nil
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, v.Type().Bits())
		if err != nil {
			return err
		}
		if v.OverflowFloat(n) {
			return errors.New("tryParse(reflect.Value.OverflowFloat)")
		}
		v.SetFloat(n)
		return nil
	case reflect.Bool:
		n, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		v.SetBool(n)
		return nil
	default:
		return fmt.Errorf("tryParse(unsupported type '%s')", v.Type().String())
	}
}

func contentType(val string) string {
	var ctype string

	if strings.ContainsRune(val, '/') {
		ctype = val
	} else {
		if !strings.HasPrefix(val, ".") {
			val = "." + val
		}
		ctype = mime.TypeByExtension(val)
	}

	return ctype
}
