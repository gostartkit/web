package web

import "testing"

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
