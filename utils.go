package web

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Redirect helper function for return url and redirect error
func Redirect(url string, code int) (any, error) {
	return nil, NewErrFn(code, "REDIRECT", func(w http.ResponseWriter, r *http.Request) error {
		http.Redirect(w, r, url, code)
		return nil
	})
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
		n, err := strconv.ParseInt(val, 10, strconv.IntSize)
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
		n, err := strconv.ParseUint(val, 10, strconv.IntSize)
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
		parts := make([]string, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			if i < 0 {
				parts = append(parts, s)
				break
			}
			parts = append(parts, s[:i])
			s = s[i+1:]
		}
		*dest = parts
		return nil
	case *[]int:
		arr := make([]int, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseInt(part, 10, strconv.IntSize)
			if err != nil {
				return err
			}
			arr = append(arr, int(n))
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]int8:
		arr := make([]int8, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseInt(part, 10, 8)
			if err != nil {
				return err
			}
			arr = append(arr, int8(n))
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]int16:
		arr := make([]int16, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseInt(part, 10, 16)
			if err != nil {
				return err
			}
			arr = append(arr, int16(n))
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]int32:
		arr := make([]int32, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseInt(part, 10, 32)
			if err != nil {
				return err
			}
			arr = append(arr, int32(n))
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]int64:
		arr := make([]int64, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return err
			}
			arr = append(arr, n)
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]uint:
		arr := make([]uint, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseUint(part, 10, 0)
			if err != nil {
				return err
			}
			arr = append(arr, uint(n))
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]uint8:
		arr := make([]uint8, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseUint(part, 10, 8)
			if err != nil {
				return err
			}
			arr = append(arr, uint8(n))
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]uint16:
		arr := make([]uint16, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseUint(part, 10, 16)
			if err != nil {
				return err
			}
			arr = append(arr, uint16(n))
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]uint32:
		arr := make([]uint32, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseUint(part, 10, 32)
			if err != nil {
				return err
			}
			arr = append(arr, uint32(n))
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]uint64:
		arr := make([]uint64, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseUint(part, 10, 64)
			if err != nil {
				return err
			}
			arr = append(arr, n)
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]float32:
		arr := make([]float32, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseFloat(part, 32)
			if err != nil {
				return err
			}
			arr = append(arr, float32(n))
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]float64:
		arr := make([]float64, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return err
			}
			arr = append(arr, n)
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]bool:
		arr := make([]bool, 0, strings.Count(val, ",")+1)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			n, err := strconv.ParseBool(part)
			if err != nil {
				return err
			}
			arr = append(arr, n)
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	default:
		return fmt.Errorf("TryParse: unsupported type %T", v)
	}
}

func TryInt(val string) (int, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(val, 10, strconv.IntSize)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func TryUint(val string) (uint, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(val, 10, strconv.IntSize)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

func TryInt8(val string) (int8, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(val, 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(n), nil
}

func TryUint8(val string) (uint8, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(val, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(n), nil
}

func TryInt16(val string) (int16, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(val, 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(n), nil
}

func TryUint16(val string) (uint16, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(val, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), nil
}

func TryInt32(val string) (int32, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}

func TryUint32(val string) (uint32, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

func TryInt64(val string) (int64, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func TryUint64(val string) (uint64, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func TryFloat32(val string) (float32, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return 0, err
	}
	return float32(n), nil
}

func TryFloat64(val string) (float64, error) {
	if val == "" {
		return 0, nil
	}
	n, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func TryBool(val string) (bool, error) {
	if val == "" {
		return false, nil
	}
	n, err := strconv.ParseBool(val)
	if err != nil {
		return false, err
	}
	return n, nil
}

func writeCode(w http.ResponseWriter, r *http.Request, code int) {
	writeCodeByMedia(w, acceptMediaType(r.Header.Get("Accept")), code)
}

func writeCodeByMedia(w http.ResponseWriter, mt mediaType, code int) {
	set := w.Header().Set

	if code == http.StatusUnauthorized {
		set("WWW-Authenticate", `Bearer realm="api", error="invalid_token", error_description="Invalid or expired token"`)
	}

	if code != http.StatusNoContent {
		set("Content-Type", contentTypeForMedia(mt))
	}

	w.WriteHeader(code)
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

func IsErrFn(err error) bool {
	_, ok := err.(*errFn)
	return ok
}
