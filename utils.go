package web

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	case *[]string:
		*dest = strings.Split(val, ",")
		return nil
	case *[]int:
		parts := strings.Split(val, ",")
		arr := make([]int, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseInt(part, 10, 0)
			if err != nil {
				return err
			}
			arr = append(arr, int(n))
		}
		*dest = arr
		return nil
	case *[]int8:
		parts := strings.Split(val, ",")
		arr := make([]int8, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseInt(part, 10, 8)
			if err != nil {
				return err
			}
			arr = append(arr, int8(n))
		}
		*dest = arr
		return nil
	case *[]int16:
		parts := strings.Split(val, ",")
		arr := make([]int16, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseInt(part, 10, 16)
			if err != nil {
				return err
			}
			arr = append(arr, int16(n))
		}
		*dest = arr
		return nil
	case *[]int32:
		parts := strings.Split(val, ",")
		arr := make([]int32, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseInt(part, 10, 32)
			if err != nil {
				return err
			}
			arr = append(arr, int32(n))
		}
		*dest = arr
		return nil
	case *[]int64:
		parts := strings.Split(val, ",")
		arr := make([]int64, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return err
			}
			arr = append(arr, n)
		}
		*dest = arr
		return nil
	case *[]uint:
		parts := strings.Split(val, ",")
		arr := make([]uint, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseUint(part, 10, 0)
			if err != nil {
				return err
			}
			arr = append(arr, uint(n))
		}
		*dest = arr
		return nil
	case *[]uint8:
		parts := strings.Split(val, ",")
		arr := make([]uint8, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseUint(part, 10, 8)
			if err != nil {
				return err
			}
			arr = append(arr, uint8(n))
		}
		*dest = arr
		return nil
	case *[]uint16:
		parts := strings.Split(val, ",")
		arr := make([]uint16, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseUint(part, 10, 16)
			if err != nil {
				return err
			}
			arr = append(arr, uint16(n))
		}
		*dest = arr
		return nil
	case *[]uint32:
		parts := strings.Split(val, ",")
		arr := make([]uint32, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseUint(part, 10, 32)
			if err != nil {
				return err
			}
			arr = append(arr, uint32(n))
		}
		*dest = arr
		return nil
	case *[]uint64:
		parts := strings.Split(val, ",")
		arr := make([]uint64, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseUint(part, 10, 64)
			if err != nil {
				return err
			}
			arr = append(arr, n)
		}
		*dest = arr
		return nil
	case *[]float32:
		parts := strings.Split(val, ",")
		arr := make([]float32, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseFloat(part, 32)
			if err != nil {
				return err
			}
			arr = append(arr, float32(n))
		}
		*dest = arr
		return nil
	case *[]float64:
		parts := strings.Split(val, ",")
		arr := make([]float64, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return err
			}
			arr = append(arr, n)
		}
		*dest = arr
		return nil
	case *[]bool:
		parts := strings.Split(val, ",")
		arr := make([]bool, 0, len(parts))
		for _, part := range parts {
			n, err := strconv.ParseBool(part)
			if err != nil {
				return err
			}
			arr = append(arr, n)
		}
		*dest = arr
		return nil
	default:
		return fmt.Errorf("TryParse: unsupported type %T", v)
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
