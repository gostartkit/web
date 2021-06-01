package web

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"reflect"
	"strconv"
	"strings"
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

	for rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	if !rv.CanSet() {
		return errors.New("TryParse(can not set value to v)")
	}

	switch rv.Interface().(type) {
	case string:
		rv.SetString(val)
		return nil
	case int, int8, int16, int32, int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		if rv.OverflowInt(n) {
			return errors.New("TryParse(reflect.Value.OverflowInt)")
		}
		rv.SetInt(n)
		return nil
	case uint, uint8, uint16, uint32, uint64, uintptr:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		if rv.OverflowUint(n) {
			return errors.New("TryParse(reflect.Value.OverflowUint)")
		}
		rv.SetUint(n)
		return nil
	case float32, float64:
		n, err := strconv.ParseFloat(val, rv.Type().Bits())
		if err != nil {
			return err
		}
		if rv.OverflowFloat(n) {
			return errors.New("TryParse(reflect.Value.OverflowFloat)")
		}
		rv.SetFloat(n)
		return nil
	case bool:
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

// binaryRead decode data from binary
func binaryRead(r io.Reader, data interface{}) error {
	return errors.New("binaryRead not implemented")
}

// binaryWrite encode data to binary
func binaryWrite(w io.Writer, data interface{}) error {
	return errors.New("binaryWrite not implemented")
}
