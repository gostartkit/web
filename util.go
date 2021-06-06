package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

const (
	maxFormSize int = 10 << 20 // 10 MB is a lot of text.
	formBufSize int = 512
)

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

// TryParse try parse val to v
func TryParse(val string, v interface{}) error {
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

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(val)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		if rv.OverflowInt(n) {
			return errors.New("TryParse(reflect.Value.OverflowInt)")
		}
		rv.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		if rv.OverflowUint(n) {
			return errors.New("TryParse(reflect.Value.OverflowUint)")
		}
		rv.SetUint(n)
		return nil
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, rv.Type().Bits())
		if err != nil {
			return err
		}
		if rv.OverflowFloat(n) {
			return errors.New("TryParse(reflect.Value.OverflowFloat)")
		}
		rv.SetFloat(n)
		return nil
	case reflect.Bool:
		n, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		rv.SetBool(n)
		return nil
	default:
		return json.Unmarshal([]byte(val), v)
	}
}

// clean clean old to new and clean space ant \t
func clean(val string, old byte, new byte) string {
	l := len(val)

	prev := 0

	var str strings.Builder
	str.Grow(l)

	for pos := 0; pos < l; pos++ {
		r := val[pos]

		switch r {
		case '/':
			if pos > prev {
				str.WriteString(val[prev:pos])
			}
			prev = pos + 1

			if pos > 0 && prev < l {
				str.WriteByte('_')
			}
		case ' ', '\t':
			if pos > prev {
				str.WriteString(val[prev:pos])
			}
			prev = pos + 1
		}
	}

	if prev < l {
		str.WriteString(val[prev:])
	}

	return str.String()
}

// binaryReader decode data from binary
func binaryReader(r io.Reader, v interface{}) error {
	return errors.New("binaryReader not implemented")
}

// binaryWriter encode data to binary
func binaryWriter(w io.Writer, v interface{}) error {
	return errors.New("binaryWriter not implemented")
}

// formReader decode data from form
// ContentType: application/x-www-form-urlencoded
func formReader(r io.Reader, v interface{}) error {
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

				if len(key) > 0 {
					if err := formKevValue(key, val, &m, &rv); err != nil {
						return err
					}
				}
				break
			}

			return err
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
	}

	return nil
}

func formKevValue(key string, value string, m *map[string]int, v *reflect.Value) error {
	log.Printf("key: %s value: %s", key, value)
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

			if err := tryParseField(value, &field); err != nil {
				return err
			}
		}
	}

	return nil
}

// formDataReader decode data from form
// ContentType: multipart/form-data
func formDataReader(r io.Reader, v interface{}) error {
	return errors.New("formDataReader not implemented")
}

// tryParseField try parse val to v
func tryParseField(val string, v *reflect.Value) error {

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		*v = v.Elem()
	}

	if !v.IsValid() {
		return errors.New("tryParseField(rv invalid)")
	}

	if !v.CanSet() {
		return errors.New("tryParseField(can not set value to rv)")
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
			return errors.New("tryParseField(reflect.Value.OverflowInt)")
		}
		v.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		if v.OverflowUint(n) {
			return errors.New("tryParseField(reflect.Value.OverflowUint)")
		}
		v.SetUint(n)
		return nil
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, v.Type().Bits())
		if err != nil {
			return err
		}
		if v.OverflowFloat(n) {
			return errors.New("tryParseField(reflect.Value.OverflowFloat)")
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
		return fmt.Errorf("tryParseField(unsupported type '%s')", v.Type().String())
	}
}
