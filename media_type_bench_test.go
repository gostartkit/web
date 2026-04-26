package web

import "testing"

func BenchmarkParseMediaTypeExactJSON(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := parseMediaType("application/json"); got != mediaJSON {
			b.Fatalf("unexpected media type: %v", got)
		}
	}
}

func BenchmarkParseMediaTypeWithParameters(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := parseMediaType("application/json; charset=utf-8"); got != mediaJSON {
			b.Fatalf("unexpected media type: %v", got)
		}
	}
}

func BenchmarkAcceptMediaTypeEmpty(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := acceptMediaType(""); got != mediaJSON {
			b.Fatalf("unexpected media type: %v", got)
		}
	}
}
