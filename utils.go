package web

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func reuseOrMakeSlice[T any](dst []T, size int) []T {
	if cap(dst) >= size {
		return dst[:0]
	}
	return make([]T, 0, size)
}

func parseUintFast64(s string) (uint64, bool) {
	if s == "" {
		return 0, false
	}

	var n uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		d := uint64(c - '0')
		if n > (math.MaxUint64-d)/10 {
			return 0, false
		}
		n = n*10 + d
	}
	return n, true
}

func parseIntFast64(s string) (int64, bool) {
	if s == "" {
		return 0, false
	}

	neg := false
	switch s[0] {
	case '-':
		neg = true
		s = s[1:]
	case '+':
		s = s[1:]
	}
	if s == "" {
		return 0, false
	}

	u, ok := parseUintFast64(s)
	if !ok {
		return 0, false
	}

	if neg {
		const maxAbsInt64 = uint64(1) << 63
		if u > maxAbsInt64 {
			return 0, false
		}
		if u == maxAbsInt64 {
			return math.MinInt64, true
		}
		return -int64(u), true
	}

	if u > math.MaxInt64 {
		return 0, false
	}
	return int64(u), true
}

func parseBoolFast(s string) (bool, bool) {
	switch s {
	case "1", "t", "T", "true", "TRUE", "True":
		return true, true
	case "0", "f", "F", "false", "FALSE", "False":
		return false, true
	default:
		return false, false
	}
}

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
		if n, ok := parseIntFast64(val); ok {
			if strconv.IntSize == 32 && (n < math.MinInt32 || n > math.MaxInt32) {
				return strconv.ErrRange
			}
			*dest = int(n)
			return nil
		}
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
		if n, ok := parseUintFast64(val); ok {
			if strconv.IntSize == 32 && n > math.MaxUint32 {
				return strconv.ErrRange
			}
			*dest = uint(n)
			return nil
		}
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
		if n, ok := parseBoolFast(val); ok {
			*dest = n
			return nil
		}
		n, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		*dest = n
		return nil
	case *[]string:
		parts := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
		arr := reuseOrMakeSlice(*dest, strings.Count(val, ",")+1)
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
	if n, ok := parseIntFast64(val); ok {
		if strconv.IntSize == 32 && (n < math.MinInt32 || n > math.MaxInt32) {
			return 0, strconv.ErrRange
		}
		return int(n), nil
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
	if n, ok := parseUintFast64(val); ok {
		if strconv.IntSize == 32 && n > math.MaxUint32 {
			return 0, strconv.ErrRange
		}
		return uint(n), nil
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
	if n, ok := parseBoolFast(val); ok {
		return n, nil
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
