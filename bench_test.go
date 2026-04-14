package web

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type benchResponseWriter struct {
	h      http.Header
	status int
}

func newBenchResponseWriter() *benchResponseWriter {
	return &benchResponseWriter{
		h: make(http.Header, 2),
	}
}

func (w *benchResponseWriter) Header() http.Header         { return w.h }
func (w *benchResponseWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w *benchResponseWriter) WriteHeader(statusCode int)  { w.status = statusCode }
func (w *benchResponseWriter) reset() {
	w.status = 0
	for k := range w.h {
		delete(w.h, k)
	}
}

func BenchmarkServeHTTPStaticJSON(b *testing.B) {
	app := New()
	out := struct {
		Ok bool `json:"ok"`
	}{Ok: true}
	app.Get("/v1/ping", func(c *Ctx) (any, error) {
		return out, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	req.Header.Set("Accept", "application/json")
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPPathParamJSON(b *testing.B) {
	app := New()
	app.Get("/v1/user/:id", func(c *Ctx) (any, error) {
		id, _ := c.ParamUint64("id")
		return map[string]uint64{"id": id}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/user/123456", nil)
	req.Header.Set("Accept", "application/json")
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkTryParseBodyJSON(b *testing.B) {
	payload := []byte(`{"id":123,"name":"sam","active":true}`)

	type reqBody struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/user", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		rec := httptest.NewRecorder()
		c := createCtx(rec, req, nil)

		var body reqBody
		if err := c.TryParseBody(&body); err != nil {
			b.Fatalf("TryParseBody failed: %v", err)
		}

		releaseCtx(c)
	}
}

func BenchmarkParamsVal(b *testing.B) {
	ps := Params{
		{Key: "tenantId", Value: "100"},
		{Key: "projectId", Value: "200"},
		{Key: "userId", Value: "300"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ps.Val("userId")
	}
}

func BenchmarkPostJSON(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	in := struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}{
		ID:     123,
		Name:   "sam",
		Active: true,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Post(b.Context(), srv.URL, "", in, nil); err != nil {
			b.Fatalf("Post failed: %v", err)
		}
	}
}

func BenchmarkServeHTTPBinary(b *testing.B) {
	app := New()
	out := []byte{0x01, 0x02, 0x03, 0x04}
	app.Get("/v1/bin", func(c *Ctx) (any, error) {
		return out, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/bin", nil)
	req.Header.Set("Accept", "application/octet-stream")
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPAvro(b *testing.B) {
	app := New()
	out := avroPayload{raw: []byte{0xAA, 0xBB, 0xCC}}
	app.Get("/v1/avro", func(c *Ctx) (any, error) {
		return out, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/avro", nil)
	req.Header.Set("Accept", "application/x-avro")
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkCtxWriteBinaryReader(b *testing.B) {
	payload := bytes.Repeat([]byte("a"), 2048)
	w := newBenchResponseWriter()
	c := &Ctx{w: w}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		if err := c.writeBinary(bytes.NewReader(payload)); err != nil {
			b.Fatalf("writeBinary failed: %v", err)
		}
	}
}

func BenchmarkTreeGetValueStatic(b *testing.B) {
	root := new(node)
	root.addRoute("/v1/ping", func(c *Ctx) (any, error) { return nil, nil })

	path := "/v1/ping"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := root.getValue(path, nil)
		if cb == nil || ps != nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v ps=%v tsr=%v", cb != nil, ps, tsr)
		}
	}
}

func BenchmarkTreeGetValueParam(b *testing.B) {
	root := new(node)
	root.addRoute("/v1/user/:id", func(c *Ctx) (any, error) { return nil, nil })
	getParams := func() *Params {
		ps := make(Params, 0, 1)
		return &ps
	}

	path := "/v1/user/123456"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := root.getValue(path, getParams)
		if cb == nil || ps == nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v psNil=%v tsr=%v", cb != nil, ps == nil, tsr)
		}
		if (*ps)[0].Key != "id" || (*ps)[0].Value != "123456" {
			b.Fatalf("unexpected param: %+v", (*ps)[0])
		}
	}
}

func BenchmarkTreeGetValueCatchAll(b *testing.B) {
	root := new(node)
	root.addRoute("/static/*filepath", func(c *Ctx) (any, error) { return nil, nil })
	getParams := func() *Params {
		ps := make(Params, 0, 1)
		return &ps
	}

	path := "/static/css/app.css"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := root.getValue(path, getParams)
		if cb == nil || ps == nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v psNil=%v tsr=%v", cb != nil, ps == nil, tsr)
		}
		if (*ps)[0].Key != "filepath" || (*ps)[0].Value != "/css/app.css" {
			b.Fatalf("unexpected param: %+v", (*ps)[0])
		}
	}
}

func BenchmarkTreeGetValueParamPooled(b *testing.B) {
	root := new(node)
	root.addRoute("/v1/user/:id", func(c *Ctx) (any, error) { return nil, nil })

	var psPool = sync.Pool{
		New: func() any {
			ps := make(Params, 0, 1)
			return &ps
		},
	}
	getParams := func() *Params {
		ps := psPool.Get().(*Params)
		*ps = (*ps)[:0]
		return ps
	}
	putParams := func(ps *Params) {
		if ps != nil {
			psPool.Put(ps)
		}
	}

	path := "/v1/user/123456"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := root.getValue(path, getParams)
		if cb == nil || ps == nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v psNil=%v tsr=%v", cb != nil, ps == nil, tsr)
		}
		if (*ps)[0].Key != "id" || (*ps)[0].Value != "123456" {
			b.Fatalf("unexpected param: %+v", (*ps)[0])
		}
		putParams(ps)
	}
}

func BenchmarkTreeGetValueCatchAllPooled(b *testing.B) {
	root := new(node)
	root.addRoute("/static/*filepath", func(c *Ctx) (any, error) { return nil, nil })

	var psPool = sync.Pool{
		New: func() any {
			ps := make(Params, 0, 1)
			return &ps
		},
	}
	getParams := func() *Params {
		ps := psPool.Get().(*Params)
		*ps = (*ps)[:0]
		return ps
	}
	putParams := func(ps *Params) {
		if ps != nil {
			psPool.Put(ps)
		}
	}

	path := "/static/css/app.css"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := root.getValue(path, getParams)
		if cb == nil || ps == nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v psNil=%v tsr=%v", cb != nil, ps == nil, tsr)
		}
		if (*ps)[0].Key != "filepath" || (*ps)[0].Value != "/css/app.css" {
			b.Fatalf("unexpected param: %+v", (*ps)[0])
		}
		putParams(ps)
	}
}
