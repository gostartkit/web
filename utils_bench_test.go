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

func BenchmarkTryBool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := TryBool("true"); err != nil {
			b.Fatalf("TryBool failed: %v", err)
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
