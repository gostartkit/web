package web

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// Redirect helper function for return url and redirect error
func Redirect(url string, code int) (Callback, error) {
	return func(w http.ResponseWriter, r *http.Request) error {
		http.Redirect(w, r, url, code)
		return nil
	}, ErrCallback
}

func ServeFile(filePath string) (Callback, error) {
	return func(w http.ResponseWriter, r *http.Request) error {
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return err
		}

		ext := filepath.Ext(filePath)

		switch ext {
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		}
		w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

		http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
		return nil
	}, ErrCallback
}

// TryParse try parse val to v
func TryParse(val string, v any) error {

	if len(val) == 0 {
		return nil
	}

	if v == nil {
		return errors.New("TryParse: nil pointer")
	}

	switch dest := v.(type) {
	case *string:
		*dest = val
		return nil
	case *int:
		n, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return err
		}
		*dest = int(n)
		return nil
	case *int8:
		n, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return err
		}
		*dest = int8(n)
		return nil
	case *int16:
		n, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return err
		}
		*dest = int16(n)
		return nil
	case *int32:
		n, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return err
		}
		*dest = int32(n)
		return nil
	case *int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		*dest = n
		return nil
	case *uint:
		n, err := strconv.ParseUint(val, 10, 0)
		if err != nil {
			return err
		}
		*dest = uint(n)
		return nil
	case *uint8:
		n, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return err
		}
		*dest = uint8(n)
		return nil
	case *uint16:
		n, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return err
		}
		*dest = uint16(n)
		return nil
	case *uint32:
		n, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return err
		}
		*dest = uint32(n)
		return nil
	case *uint64:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		*dest = n
		return nil
	case *float32:
		n, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return err
		}
		*dest = float32(n)
		return nil
	case *float64:
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		*dest = n
		return nil
	case *bool:
		n, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		*dest = n
		return nil
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			return fmt.Errorf("TryParse: non-pointer %s", reflect.TypeOf(v).String())
		}
		if rv.IsNil() {
			return errors.New("TryParse: nil pointer")
		}
		rv = rv.Elem()
		if !rv.CanSet() {
			return errors.New("TryParse: cannot set value")
		}
		return tryParse(val, &rv)
	}

}

// tryParse try parse val to v
func tryParse(val string, v *reflect.Value) error {

	retry := 3

	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		*v = v.Elem()
		if retry--; retry < 0 {
			return errors.New("tryParse: invalid pointer")
		}
	}

	if !v.IsValid() {
		return errors.New("tryParse: invalid value")
	}

	if !v.CanSet() {
		return errors.New("tryParse: unsettable value")
	}

	switch v.Interface().(type) {
	case string:
		v.SetString(val)
		return nil
	case int:
		n, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return err
		}
		v.SetInt(n)
		return nil
	case int8:
		n, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return err
		}
		v.SetInt(n)
		return nil
	case int16:
		n, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return err
		}
		v.SetInt(n)
		return nil
	case int32:
		n, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return err
		}
		v.SetInt(n)
		return nil
	case int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(n)
		return nil
	case uint:
		n, err := strconv.ParseUint(val, 10, 0)
		if err != nil {
			return err
		}
		v.SetUint(n)
		return nil
	case uint8:
		n, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return err
		}
		v.SetUint(n)
		return nil
	case uint16:
		n, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return err
		}
		v.SetUint(n)
		return nil
	case uint32:
		n, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return err
		}
		v.SetUint(n)
		return nil
	case uint64:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(n)
		return nil
	case float32:
		n, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return err
		}
		v.SetFloat(n)
		return nil
	case float64:
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		v.SetFloat(n)
		return nil
	case bool:
		n, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		v.SetBool(n)
		return nil
	default:
		return fmt.Errorf("tryParse: unsupported type '%s'", v.Type().String())
	}

}

// bearerToken return token
func bearerToken(auth string) string {
	const prefix = "Bearer "
	l := len(prefix)

	if len(auth) < l || !strings.EqualFold(auth[:l], prefix) {
		return ""
	}

	return auth[l:]
}
