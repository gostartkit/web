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
	}, ErrCallBack
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
	}, ErrCallBack
}

// TryParse try parse val to v
func TryParse(val string, v any) error {

	if len(val) == 0 {
		return nil
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

// bearerToken return token
func bearerToken(auth string) string {
	const prefix = "Bearer "
	l := len(prefix)

	if len(auth) < l || !strings.EqualFold(auth[:l], prefix) {
		return ""
	}

	return auth[l:]
}
