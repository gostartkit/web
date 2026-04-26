package web

import "testing"

func BenchmarkTryInt(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := TryInt("123456"); err != nil {
			b.Fatalf("TryInt failed: %v", err)
		}
	}
}

func BenchmarkTryUint(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := TryUint("123456"); err != nil {
			b.Fatalf("TryUint failed: %v", err)
		}
	}
}

func BenchmarkTryInt64(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := TryInt64("1234567890123"); err != nil {
			b.Fatalf("TryInt64 failed: %v", err)
		}
	}
}

func BenchmarkTryUint64(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := TryUint64("1234567890123"); err != nil {
			b.Fatalf("TryUint64 failed: %v", err)
		}
	}
}

func BenchmarkTryBool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := TryBool("true"); err != nil {
			b.Fatalf("TryBool failed: %v", err)
		}
	}
}

func BenchmarkTryParseInt64(b *testing.B) {
	var out int64

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := TryParse("1234567890123", &out); err != nil {
			b.Fatalf("TryParse failed: %v", err)
		}
	}
}

func BenchmarkTryParseUint64(b *testing.B) {
	var out uint64

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := TryParse("1234567890123", &out); err != nil {
			b.Fatalf("TryParse failed: %v", err)
		}
	}
}

func BenchmarkTryParseInt8(b *testing.B) {
	var out int8

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := TryParse("123", &out); err != nil {
			b.Fatalf("TryParse failed: %v", err)
		}
	}
}

func BenchmarkTryParseUint8(b *testing.B) {
	var out uint8

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := TryParse("123", &out); err != nil {
			b.Fatalf("TryParse failed: %v", err)
		}
	}
}

func BenchmarkTryParseInt16(b *testing.B) {
	var out int16

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := TryParse("12345", &out); err != nil {
			b.Fatalf("TryParse failed: %v", err)
		}
	}
}

func BenchmarkTryParseUint16(b *testing.B) {
	var out uint16

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := TryParse("12345", &out); err != nil {
			b.Fatalf("TryParse failed: %v", err)
		}
	}
}

func BenchmarkTryParseInt32(b *testing.B) {
	var out int32

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := TryParse("1234567890", &out); err != nil {
			b.Fatalf("TryParse failed: %v", err)
		}
	}
}

func BenchmarkTryParseUint32(b *testing.B) {
	var out uint32

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := TryParse("1234567890", &out); err != nil {
			b.Fatalf("TryParse failed: %v", err)
		}
	}
}

func BenchmarkTryParseIntSlice(b *testing.B) {
	input := "1,2,3,4,5,6,7,8,9,10"
	var out []int

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out = out[:0]
		if err := TryParse(input, &out); err != nil {
			b.Fatalf("TryParse failed: %v", err)
		}
	}
}

func BenchmarkTryParseStringSlice(b *testing.B) {
	input := "alpha,beta,gamma,delta,epsilon"
	var out []string

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out = out[:0]
		if err := TryParse(input, &out); err != nil {
			b.Fatalf("TryParse failed: %v", err)
		}
	}
}
