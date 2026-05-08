package web

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func reuseOrMakeCommaSlice[T any](dst []T, val string) []T {
	if cap(dst) > 0 {
		return dst[:0]
	}
	return make([]T, 0, strings.Count(val, ",")+1)
}

func parseIntCSVFast(val string, dst []int) ([]int, error) {
	arr := reuseOrMakeCommaSlice(dst, val)
	maxAbs := uint64(math.MaxInt64)
	negMaxAbs := uint64(1) << 63
	if strconv.IntSize == 32 {
		maxAbs = math.MaxInt32
		negMaxAbs = uint64(1) << 31
	}

	for i := 0; i < len(val); {
		start := i
		neg := false
		switch val[i] {
		case '-':
			neg = true
			i++
		case '+':
			i++
		}

		limit := maxAbs
		if neg {
			limit = negMaxAbs
		}

		digitsStart := i
		var n uint64
		for i < len(val) && val[i] != ',' {
			c := val[i]
			if c < '0' || c > '9' {
				end := i
				for end < len(val) && val[end] != ',' {
					end++
				}
				_, err := strconv.ParseInt(val[start:end], 10, strconv.IntSize)
				return arr, err
			}
			d := uint64(c - '0')
			if n > (limit-d)/10 {
				return arr, strconv.ErrRange
			}
			n = n*10 + d
			i++
		}

		if i == digitsStart {
			_, err := strconv.ParseInt(val[start:i], 10, strconv.IntSize)
			return arr, err
		}

		if neg {
			if strconv.IntSize == 64 && n == negMaxAbs {
				arr = append(arr, int(n))
			} else {
				arr = append(arr, int(-int64(n)))
			}
		} else {
			arr = append(arr, int(n))
		}

		if i == len(val) {
			break
		}
		i++
		if i == len(val) {
			_, err := strconv.ParseInt("", 10, strconv.IntSize)
			return arr, err
		}
	}

	return arr, nil
}

func parseUintFast64(s string) (uint64, bool) {
	if s == "" {
		return 0, false
	}

	var n uint64
	if len(s) < 20 {
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c < '0' || c > '9' {
				return 0, false
			}
			n = n*10 + uint64(c-'0')
		}
		return n, true
	}

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

// Redirect returns a handler result that performs an HTTP redirect.
//
// Use it directly from a route handler:
//
//	return web.Redirect("/login", http.StatusFound)
//
// The returned error carries code and writes through http.Redirect on the
// framework error path.
//
// Deprecated: prefer c.Redirect(statusCode, url) from a handler:
//
//	return nil, c.Redirect(http.StatusFound, "/login")
func Redirect(url string, code int) (any, error) {
	return nil, NewErrFn(code, "REDIRECT", func(w http.ResponseWriter, r *http.Request) error {
		http.Redirect(w, r, url, code)
		return nil
	})
}

// TryParse parses val into the destination pointer v.
//
// Supported destinations include pointers to string, signed and unsigned integer
// widths, float32, float64, bool, and comma-separated slices of those scalar
// types. Empty val is treated as "not provided": TryParse returns nil and leaves
// the destination unchanged.
//
// Example:
//
//	var ids []uint64
//	if err := web.TryParse("1,2,3", &ids); err != nil {
//		return nil, err
//	}
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
		if n, ok := parseIntFast64(val); ok {
			if n < math.MinInt8 || n > math.MaxInt8 {
				return strconv.ErrRange
			}
			*dest = int8(n)
			return nil
		}
		n, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return err
		}
		*dest = int8(n)
		return nil
	case *int16:
		if n, ok := parseIntFast64(val); ok {
			if n < math.MinInt16 || n > math.MaxInt16 {
				return strconv.ErrRange
			}
			*dest = int16(n)
			return nil
		}
		n, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return err
		}
		*dest = int16(n)
		return nil
	case *int32:
		if n, ok := parseIntFast64(val); ok {
			if n < math.MinInt32 || n > math.MaxInt32 {
				return strconv.ErrRange
			}
			*dest = int32(n)
			return nil
		}
		n, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return err
		}
		*dest = int32(n)
		return nil
	case *int64:
		if n, ok := parseIntFast64(val); ok {
			*dest = n
			return nil
		}
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
		if n, ok := parseUintFast64(val); ok {
			if n > math.MaxUint8 {
				return strconv.ErrRange
			}
			*dest = uint8(n)
			return nil
		}
		n, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return err
		}
		*dest = uint8(n)
		return nil
	case *uint16:
		if n, ok := parseUintFast64(val); ok {
			if n > math.MaxUint16 {
				return strconv.ErrRange
			}
			*dest = uint16(n)
			return nil
		}
		n, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return err
		}
		*dest = uint16(n)
		return nil
	case *uint32:
		if n, ok := parseUintFast64(val); ok {
			if n > math.MaxUint32 {
				return strconv.ErrRange
			}
			*dest = uint32(n)
			return nil
		}
		n, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return err
		}
		*dest = uint32(n)
		return nil
	case *uint64:
		if n, ok := parseUintFast64(val); ok {
			*dest = n
			return nil
		}
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
		parts := reuseOrMakeCommaSlice(*dest, val)
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
		arr, err := parseIntCSVFast(val, *dest)
		if err != nil {
			return err
		}
		*dest = arr
		return nil
	case *[]int8:
		arr := reuseOrMakeCommaSlice(*dest, val)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			if n, ok := parseIntFast64(part); ok {
				if n < math.MinInt8 || n > math.MaxInt8 {
					return strconv.ErrRange
				}
				arr = append(arr, int8(n))
			} else {
				n, err := strconv.ParseInt(part, 10, 8)
				if err != nil {
					return err
				}
				arr = append(arr, int8(n))
			}
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]int16:
		arr := reuseOrMakeCommaSlice(*dest, val)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			if n, ok := parseIntFast64(part); ok {
				if n < math.MinInt16 || n > math.MaxInt16 {
					return strconv.ErrRange
				}
				arr = append(arr, int16(n))
			} else {
				n, err := strconv.ParseInt(part, 10, 16)
				if err != nil {
					return err
				}
				arr = append(arr, int16(n))
			}
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]int32:
		arr := reuseOrMakeCommaSlice(*dest, val)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			if n, ok := parseIntFast64(part); ok {
				if n < math.MinInt32 || n > math.MaxInt32 {
					return strconv.ErrRange
				}
				arr = append(arr, int32(n))
			} else {
				n, err := strconv.ParseInt(part, 10, 32)
				if err != nil {
					return err
				}
				arr = append(arr, int32(n))
			}
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]int64:
		arr := reuseOrMakeCommaSlice(*dest, val)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			if n, ok := parseIntFast64(part); ok {
				arr = append(arr, n)
			} else {
				n, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					return err
				}
				arr = append(arr, n)
			}
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]uint:
		arr := reuseOrMakeCommaSlice(*dest, val)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			if n, ok := parseUintFast64(part); ok {
				if strconv.IntSize == 32 && n > math.MaxUint32 {
					return strconv.ErrRange
				}
				arr = append(arr, uint(n))
			} else {
				n, err := strconv.ParseUint(part, 10, 0)
				if err != nil {
					return err
				}
				arr = append(arr, uint(n))
			}
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]uint8:
		arr := reuseOrMakeCommaSlice(*dest, val)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			if n, ok := parseUintFast64(part); ok {
				if n > math.MaxUint8 {
					return strconv.ErrRange
				}
				arr = append(arr, uint8(n))
			} else {
				n, err := strconv.ParseUint(part, 10, 8)
				if err != nil {
					return err
				}
				arr = append(arr, uint8(n))
			}
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]uint16:
		arr := reuseOrMakeCommaSlice(*dest, val)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			if n, ok := parseUintFast64(part); ok {
				if n > math.MaxUint16 {
					return strconv.ErrRange
				}
				arr = append(arr, uint16(n))
			} else {
				n, err := strconv.ParseUint(part, 10, 16)
				if err != nil {
					return err
				}
				arr = append(arr, uint16(n))
			}
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]uint32:
		arr := reuseOrMakeCommaSlice(*dest, val)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			if n, ok := parseUintFast64(part); ok {
				if n > math.MaxUint32 {
					return strconv.ErrRange
				}
				arr = append(arr, uint32(n))
			} else {
				n, err := strconv.ParseUint(part, 10, 32)
				if err != nil {
					return err
				}
				arr = append(arr, uint32(n))
			}
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]uint64:
		arr := reuseOrMakeCommaSlice(*dest, val)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			if n, ok := parseUintFast64(part); ok {
				arr = append(arr, n)
			} else {
				n, err := strconv.ParseUint(part, 10, 64)
				if err != nil {
					return err
				}
				arr = append(arr, n)
			}
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
		*dest = arr
		return nil
	case *[]float32:
		arr := reuseOrMakeCommaSlice(*dest, val)
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
		arr := reuseOrMakeCommaSlice(*dest, val)
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
		arr := reuseOrMakeCommaSlice(*dest, val)
		s := val
		for {
			i := strings.IndexByte(s, ',')
			part := s
			if i >= 0 {
				part = s[:i]
			}
			if n, ok := parseBoolFast(part); ok {
				arr = append(arr, n)
			} else {
				n, err := strconv.ParseBool(part)
				if err != nil {
					return err
				}
				arr = append(arr, n)
			}
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

// TryInt parses val as int.
//
// Empty strings return zero and nil, matching the framework's optional
// parameter/query behavior.
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

// TryUint parses val as uint.
//
// Empty strings return zero and nil. Overflow is reported as strconv.ErrRange.
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

// TryInt8 parses val as int8.
//
// Empty strings return zero and nil. Values outside the int8 range return
// strconv.ErrRange.
func TryInt8(val string) (int8, error) {
	if val == "" {
		return 0, nil
	}
	if n, ok := parseIntFast64(val); ok {
		if n < math.MinInt8 || n > math.MaxInt8 {
			return 0, strconv.ErrRange
		}
		return int8(n), nil
	}
	n, err := strconv.ParseInt(val, 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(n), nil
}

// TryUint8 parses val as uint8.
//
// Empty strings return zero and nil. Values above math.MaxUint8 return
// strconv.ErrRange.
func TryUint8(val string) (uint8, error) {
	if val == "" {
		return 0, nil
	}
	if n, ok := parseUintFast64(val); ok {
		if n > math.MaxUint8 {
			return 0, strconv.ErrRange
		}
		return uint8(n), nil
	}
	n, err := strconv.ParseUint(val, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(n), nil
}

// TryInt16 parses val as int16.
//
// Empty strings return zero and nil. Values outside the int16 range return
// strconv.ErrRange.
func TryInt16(val string) (int16, error) {
	if val == "" {
		return 0, nil
	}
	if n, ok := parseIntFast64(val); ok {
		if n < math.MinInt16 || n > math.MaxInt16 {
			return 0, strconv.ErrRange
		}
		return int16(n), nil
	}
	n, err := strconv.ParseInt(val, 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(n), nil
}

// TryUint16 parses val as uint16.
//
// Empty strings return zero and nil. Values above math.MaxUint16 return
// strconv.ErrRange.
func TryUint16(val string) (uint16, error) {
	if val == "" {
		return 0, nil
	}
	if n, ok := parseUintFast64(val); ok {
		if n > math.MaxUint16 {
			return 0, strconv.ErrRange
		}
		return uint16(n), nil
	}
	n, err := strconv.ParseUint(val, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), nil
}

// TryInt32 parses val as int32.
//
// Empty strings return zero and nil. Values outside the int32 range return
// strconv.ErrRange.
func TryInt32(val string) (int32, error) {
	if val == "" {
		return 0, nil
	}
	if n, ok := parseIntFast64(val); ok {
		if n < math.MinInt32 || n > math.MaxInt32 {
			return 0, strconv.ErrRange
		}
		return int32(n), nil
	}
	n, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}

// TryUint32 parses val as uint32.
//
// Empty strings return zero and nil. Values above math.MaxUint32 return
// strconv.ErrRange.
func TryUint32(val string) (uint32, error) {
	if val == "" {
		return 0, nil
	}
	if n, ok := parseUintFast64(val); ok {
		if n > math.MaxUint32 {
			return 0, strconv.ErrRange
		}
		return uint32(n), nil
	}
	n, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// TryInt64 parses val as int64.
//
// Empty strings return zero and nil. The hot path avoids allocations for common
// decimal input and falls back to strconv for standard error reporting.
func TryInt64(val string) (int64, error) {
	if val == "" {
		return 0, nil
	}
	if n, ok := parseIntFast64(val); ok {
		return n, nil
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// TryUint64 parses val as uint64.
//
// Empty strings return zero and nil. This helper is used by Ctx.ParamUint64 and
// QueryUint64 for fast numeric ID parsing.
func TryUint64(val string) (uint64, error) {
	if val == "" {
		return 0, nil
	}
	if n, ok := parseUintFast64(val); ok {
		return n, nil
	}
	n, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// TryFloat32 parses val as float32.
//
// Empty strings return zero and nil. Non-empty input follows strconv.ParseFloat
// semantics.
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

// TryFloat64 parses val as float64.
//
// Empty strings return zero and nil. Non-empty input follows strconv.ParseFloat
// semantics.
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

// TryBool parses val as bool.
//
// Empty strings return false and nil. Common boolean spellings such as "true",
// "false", "1", and "0" are handled by a fast path before falling back to
// strconv.ParseBool.
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
	header := w.Header()

	if code == http.StatusUnauthorized {
		header.Set("WWW-Authenticate", `Bearer realm="api", error="invalid_token", error_description="Invalid or expired token"`)
	}

	if code != http.StatusNoContent {
		header.Set("Content-Type", contentTypeForMedia(mt))
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

// IsErrFn reports whether err carries a custom response callback created by NewErrFn.
//
// It is primarily used internally by the framework error path, but can help
// tests distinguish redirect/custom-response errors from ordinary framework
// errors.
func IsErrFn(err error) bool {
	_, ok := err.(*errFn)
	return ok
}
