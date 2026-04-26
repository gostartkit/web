package web

import (
	"errors"
	"math"
	"strconv"
	"testing"
)

func TestTryParseIntegerScalarFastPaths(t *testing.T) {
	t.Parallel()

	t.Run("int8", func(t *testing.T) {
		var v int8
		if err := TryParse("+127", &v); err != nil {
			t.Fatalf("TryParse int8 max: %v", err)
		}
		if v != math.MaxInt8 {
			t.Fatalf("expected %d, got %d", math.MaxInt8, v)
		}
		if err := TryParse("-128", &v); err != nil {
			t.Fatalf("TryParse int8 min: %v", err)
		}
		if v != math.MinInt8 {
			t.Fatalf("expected %d, got %d", math.MinInt8, v)
		}
		if err := TryParse("128", &v); !errors.Is(err, strconv.ErrRange) {
			t.Fatalf("expected range error, got %v", err)
		}
	})

	t.Run("int16", func(t *testing.T) {
		var v int16
		if err := TryParse("32767", &v); err != nil {
			t.Fatalf("TryParse int16 max: %v", err)
		}
		if v != math.MaxInt16 {
			t.Fatalf("expected %d, got %d", math.MaxInt16, v)
		}
		if err := TryParse("-32768", &v); err != nil {
			t.Fatalf("TryParse int16 min: %v", err)
		}
		if v != math.MinInt16 {
			t.Fatalf("expected %d, got %d", math.MinInt16, v)
		}
		if err := TryParse("32768", &v); !errors.Is(err, strconv.ErrRange) {
			t.Fatalf("expected range error, got %v", err)
		}
	})

	t.Run("int32", func(t *testing.T) {
		var v int32
		if err := TryParse("2147483647", &v); err != nil {
			t.Fatalf("TryParse int32 max: %v", err)
		}
		if v != math.MaxInt32 {
			t.Fatalf("expected %d, got %d", math.MaxInt32, v)
		}
		if err := TryParse("-2147483648", &v); err != nil {
			t.Fatalf("TryParse int32 min: %v", err)
		}
		if v != math.MinInt32 {
			t.Fatalf("expected %d, got %d", math.MinInt32, v)
		}
		if err := TryParse("2147483648", &v); !errors.Is(err, strconv.ErrRange) {
			t.Fatalf("expected range error, got %v", err)
		}
	})

	t.Run("int64", func(t *testing.T) {
		var v int64
		if err := TryParse("9223372036854775807", &v); err != nil {
			t.Fatalf("TryParse int64 max: %v", err)
		}
		if v != math.MaxInt64 {
			t.Fatalf("expected %d, got %d", int64(math.MaxInt64), v)
		}
		if err := TryParse("-9223372036854775808", &v); err != nil {
			t.Fatalf("TryParse int64 min: %v", err)
		}
		if v != math.MinInt64 {
			t.Fatalf("expected %d, got %d", int64(math.MinInt64), v)
		}
		if err := TryParse("9223372036854775808", &v); !errors.Is(err, strconv.ErrRange) {
			t.Fatalf("expected range error, got %v", err)
		}
	})

	t.Run("uint8", func(t *testing.T) {
		var v uint8
		if err := TryParse("255", &v); err != nil {
			t.Fatalf("TryParse uint8 max: %v", err)
		}
		if v != math.MaxUint8 {
			t.Fatalf("expected %d, got %d", math.MaxUint8, v)
		}
		if err := TryParse("256", &v); !errors.Is(err, strconv.ErrRange) {
			t.Fatalf("expected range error, got %v", err)
		}
		if err := TryParse("-1", &v); err == nil {
			t.Fatalf("expected error for negative uint8")
		}
	})

	t.Run("uint16", func(t *testing.T) {
		var v uint16
		if err := TryParse("65535", &v); err != nil {
			t.Fatalf("TryParse uint16 max: %v", err)
		}
		if v != math.MaxUint16 {
			t.Fatalf("expected %d, got %d", math.MaxUint16, v)
		}
		if err := TryParse("65536", &v); !errors.Is(err, strconv.ErrRange) {
			t.Fatalf("expected range error, got %v", err)
		}
	})

	t.Run("uint32", func(t *testing.T) {
		var v uint32
		if err := TryParse("4294967295", &v); err != nil {
			t.Fatalf("TryParse uint32 max: %v", err)
		}
		if v != math.MaxUint32 {
			t.Fatalf("expected %d, got %d", uint32(math.MaxUint32), v)
		}
		if err := TryParse("4294967296", &v); !errors.Is(err, strconv.ErrRange) {
			t.Fatalf("expected range error, got %v", err)
		}
	})

	t.Run("uint64", func(t *testing.T) {
		var v uint64
		if err := TryParse("18446744073709551615", &v); err != nil {
			t.Fatalf("TryParse uint64 max: %v", err)
		}
		if v != math.MaxUint64 {
			t.Fatalf("expected %d, got %d", uint64(math.MaxUint64), v)
		}
		if err := TryParse("18446744073709551616", &v); !errors.Is(err, strconv.ErrRange) {
			t.Fatalf("expected range error, got %v", err)
		}
	})
}
