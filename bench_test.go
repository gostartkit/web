package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"unsafe"
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
	type out struct {
		ID uint64 `json:"id"`
	}
	app.Get("/v1/user/:id", func(c *Ctx) (any, error) {
		id, _ := c.ParamUint64("id")
		return out{ID: id}, nil
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

func BenchmarkServeHTTPStaticOverParamSiblingJSON(b *testing.B) {
	app := New()
	type out struct {
		ID string `json:"id"`
	}
	app.Get("/organizations/:id/devices/:device_id", func(c *Ctx) (any, error) {
		return out{ID: c.Param("device_id")}, nil
	})
	app.Get("/organizations/:id/devices/provision", func(c *Ctx) (any, error) {
		return out{ID: c.Param("id")}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/organizations/123/devices/provision", nil)
	req.Header.Set("Accept", "application/json")
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPStaticOverParamSiblingNoContent(b *testing.B) {
	app := New()
	app.Get("/organizations/:id/devices/:device_id", func(c *Ctx) (any, error) {
		if c.Param("device_id") == "" {
			b.Fatal("missing device_id")
		}
		return nil, nil
	})
	app.Get("/organizations/:id/devices/provision", func(c *Ctx) (any, error) {
		if c.Param("id") == "" {
			b.Fatal("missing id")
		}
		return nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/organizations/123/devices/provision", nil)
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPStaticJSONRawMessage(b *testing.B) {
	app := New()
	out := json.RawMessage(`{"ok":true}`)
	app.Get("/v1/raw", func(c *Ctx) (any, error) {
		return out, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/raw", nil)
	req.Header.Set("Accept", "application/json")
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPNoContent(b *testing.B) {
	app := New()
	app.Get("/v1/empty", func(c *Ctx) (any, error) {
		return nil, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/empty", nil)
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPManualWrite(b *testing.B) {
	app := New()
	payload := []byte(`{"ok":true}`)
	app.Get("/v1/manual", func(c *Ctx) (any, error) {
		_, err := c.Write(payload)
		return nil, err
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/manual", nil)
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPStandardHandler(b *testing.B) {
	app := New()
	payload := []byte(`{"ok":true}`)
	app.GetHTTP("/v1/std", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/std", nil)
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPBlob(b *testing.B) {
	app := New()
	payload := []byte(`{"ok":true}`)
	app.Get("/v1/blob", func(c *Ctx) (any, error) {
		return nil, c.Blob(http.StatusOK, "application/octet-stream", payload)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/blob", nil)
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPCustomJSONWriter(b *testing.B) {
	app := New()
	payload := []byte(`{"ok":true}`)
	if err := app.RegisterWriter("application/json", func(c *Ctx, v any) error {
		_, err := c.Write(payload)
		return err
	}); err != nil {
		b.Fatalf("RegisterWriter failed: %v", err)
	}
	app.Get("/v1/custom", func(c *Ctx) (any, error) {
		return payload, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/custom", nil)
	req.Header.Set("Accept", "application/json")
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkRouteRegistrationStaticParamSiblings(b *testing.B) {
	paths := make([]string, 128)
	for i := range paths {
		paths[i] = "/organizations/:id/devices/static-" + strconv.Itoa(i)
	}
	handler := func(c *Ctx) (any, error) { return nil, nil }

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app := New()
		app.Get("/organizations/:id/devices/:device_id", handler)
		for _, path := range paths {
			app.Get(path, handler)
		}
		app.Finalize()
	}
}

func BenchmarkServeHTTPMethodNotAllowedStatic(b *testing.B) {
	app := New(WithMethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})))
	app.Get("/users", func(c *Ctx) (any, error) { return nil, nil })
	app.Finalize()

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	w := newBenchResponseWriter()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		app.ServeHTTP(w, req)
	}
}

func BenchmarkServeHTTPMethodNotAllowedParam(b *testing.B) {
	app := New(WithMethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})))
	app.Get("/users/:id", func(c *Ctx) (any, error) { return nil, nil })
	app.Finalize()

	req := httptest.NewRequest(http.MethodPost, "/users/123", nil)
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
		c := createCtx(nil, rec, req, nil)

		var body reqBody
		if err := c.TryParseBody(&body); err != nil {
			b.Fatalf("TryParseBody failed: %v", err)
		}

		releaseCtx(c)
	}
}

func BenchmarkTryParseJSONBodyFast(b *testing.B) {
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
		rec := httptest.NewRecorder()
		c := createCtx(nil, rec, req, nil)

		var body reqBody
		if err := c.TryParseJSONBodyFast(&body); err != nil {
			b.Fatalf("TryParseJSONBodyFast failed: %v", err)
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

func BenchmarkCtxCreateRelease(b *testing.B) {
	app := New()
	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := createCtx(app, w, req, nil)
		releaseCtx(c)
	}
	b.ReportMetric(float64(unsafe.Sizeof(Ctx{})), "bytes/ctx")
}

func BenchmarkCtxCreateReleaseRouteMatch(b *testing.B) {
	app := New()
	req := httptest.NewRequest(http.MethodGet, "/v1/user/123456", nil)
	w := newBenchResponseWriter()
	match := routeMatch{
		params: routeParams{
			names:  []string{"id"},
			count:  1,
			value0: "123456",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := createCtxWithRouteMatch(app, w, req, &match)
		releaseCtx(c)
	}
	b.ReportMetric(float64(unsafe.Sizeof(Ctx{})), "bytes/ctx")
}

func BenchmarkCtxParamUint64(b *testing.B) {
	ps := Params{
		{Key: "id", Value: "1234567890123"},
	}
	c := &Ctx{param: &ps}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.ParamUint64("id"); err != nil {
			b.Fatalf("ParamUint64 failed: %v", err)
		}
	}
}

func BenchmarkCtxParamUint64RouteMatch(b *testing.B) {
	rc := &pooledRouteCtx{
		params: routeParams{
			names:  []string{"id"},
			count:  1,
			value0: "1234567890123",
		},
	}
	c := &rc.Ctx
	c.routeCtx = rc

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.ParamUint64("id"); err != nil {
			b.Fatalf("ParamUint64 failed: %v", err)
		}
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
		if err := Post(context.Background(), srv.URL, "", in, nil); err != nil {
			b.Fatalf("Post failed: %v", err)
		}
	}
}

func BenchmarkPostBytes(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	in := []byte(`{"id":123,"name":"sam","active":true}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := PostBytes(context.Background(), srv.URL, "", in, nil); err != nil {
			b.Fatalf("PostBytes failed: %v", err)
		}
	}
}

func BenchmarkDoReqWithClientStruct(b *testing.B) {
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		b.Fatalf("new request failed: %v", err)
	}

	var out struct {
		Ok bool `json:"ok"`
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out.Ok = false
		if err := DoReqWithClient(client, req, &out, nil); err != nil {
			b.Fatalf("DoReqWithClient failed: %v", err)
		}
	}
}

func BenchmarkDoReqWithClientRawBody(b *testing.B) {
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		b.Fatalf("new request failed: %v", err)
	}

	var out RawBody

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out = out[:0]
		if err := DoReqWithClient(client, req, &out, nil); err != nil {
			b.Fatalf("DoReqWithClient failed: %v", err)
		}
	}
}

func BenchmarkDoReqWithClientRawBody64KB(b *testing.B) {
	payload := bytes.Repeat([]byte("a"), 64*1024)
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(payload)),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		b.Fatalf("new request failed: %v", err)
	}

	out := make(RawBody, 0, len(payload)+bytes.MinRead)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out = out[:0]
		if err := DoReqWithClient(client, req, &out, nil); err != nil {
			b.Fatalf("DoReqWithClient failed: %v", err)
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

func BenchmarkCtxWriteJSONRawMessage(b *testing.B) {
	payload := json.RawMessage(`{"ok":true}`)
	w := newBenchResponseWriter()
	c := &Ctx{w: w}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		if err := c.writeJSON(payload); err != nil {
			b.Fatalf("writeJSON failed: %v", err)
		}
	}
}

func BenchmarkCtxWriteBinaryBytes(b *testing.B) {
	payload := []byte(`{"ok":true}`)
	w := newBenchResponseWriter()
	c := &Ctx{w: w}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		if err := c.writeBinary(payload); err != nil {
			b.Fatalf("writeBinary failed: %v", err)
		}
	}
}

func BenchmarkCtxWriteBinaryString(b *testing.B) {
	payload := `{"ok":true}`
	w := newBenchResponseWriter()
	c := &Ctx{w: w}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		if err := c.writeBinary(payload); err != nil {
			b.Fatalf("writeBinary failed: %v", err)
		}
	}
}

func BenchmarkCtxWriteAvroMarshaler(b *testing.B) {
	payload := avroPayload{raw: []byte{0xAA, 0xBB, 0xCC}}
	w := newBenchResponseWriter()
	c := &Ctx{w: w}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.reset()
		if err := c.writeAvro(payload); err != nil {
			b.Fatalf("writeAvro failed: %v", err)
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
		cb, ps, tsr := root.getValueFast(path, nil)
		if cb == nil || ps != nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v ps=%v tsr=%v", cb != nil, ps, tsr)
		}
	}
}

func BenchmarkTreeGetValueParam(b *testing.B) {
	root := new(node)
	root.addRoute("/v1/user/:id", func(c *Ctx) (any, error) { return nil, nil })
	app := New()
	app.maxParams = 1

	path := "/v1/user/123456"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := root.getValueFast(path, app)
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
	app := New()
	app.maxParams = 1

	path := "/static/css/app.css"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := root.getValueFast(path, app)
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
	app := New()
	app.maxParams = 1

	path := "/v1/user/123456"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := root.getValueFast(path, app)
		if cb == nil || ps == nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v psNil=%v tsr=%v", cb != nil, ps == nil, tsr)
		}
		if (*ps)[0].Key != "id" || (*ps)[0].Value != "123456" {
			b.Fatalf("unexpected param: %+v", (*ps)[0])
		}
		app.putParams(ps)
	}
}

func BenchmarkTreeGetValueStaticOverParamSiblingPooled(b *testing.B) {
	root := new(node)
	root.addRoute("/organizations/:id/devices/:device_id", func(c *Ctx) (any, error) { return nil, nil })
	root.addRoute("/organizations/:id/devices/provision", func(c *Ctx) (any, error) { return nil, nil })
	frozen := root.freeze()
	app := New()
	app.maxParams = 2

	path := "/organizations/123/devices/provision"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := frozen.getValue(path, app)
		if cb == nil || ps == nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v psNil=%v tsr=%v", cb != nil, ps == nil, tsr)
		}
		if len(*ps) != 1 || (*ps)[0].Key != "id" || (*ps)[0].Value != "123" {
			b.Fatalf("unexpected params: %+v", *ps)
		}
		app.putParams(ps)
	}
}

func BenchmarkTreeGetValueParamWithStaticSiblingPooled(b *testing.B) {
	root := new(node)
	root.addRoute("/organizations/:id/devices/:device_id", func(c *Ctx) (any, error) { return nil, nil })
	root.addRoute("/organizations/:id/devices/provision", func(c *Ctx) (any, error) { return nil, nil })
	frozen := root.freeze()
	app := New()
	app.maxParams = 2

	path := "/organizations/123/devices/456"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := frozen.getValue(path, app)
		if cb == nil || ps == nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v psNil=%v tsr=%v", cb != nil, ps == nil, tsr)
		}
		if len(*ps) != 2 || (*ps)[0].Key != "id" || (*ps)[0].Value != "123" || (*ps)[1].Key != "device_id" || (*ps)[1].Value != "456" {
			b.Fatalf("unexpected params: %+v", *ps)
		}
		app.putParams(ps)
	}
}

func BenchmarkTreeGetValueParamAfterStaticPrefixMissPooled(b *testing.B) {
	root := new(node)
	root.addRoute("/organizations/:id/devices/:device_id", func(c *Ctx) (any, error) { return nil, nil })
	root.addRoute("/organizations/:id/devices/config/rollout", func(c *Ctx) (any, error) { return nil, nil })
	frozen := root.freeze()
	app := New()
	app.maxParams = 2

	path := "/organizations/123/devices/configured"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := frozen.getValue(path, app)
		if cb == nil || ps == nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v psNil=%v tsr=%v", cb != nil, ps == nil, tsr)
		}
		if len(*ps) != 2 || (*ps)[0].Key != "id" || (*ps)[0].Value != "123" || (*ps)[1].Key != "device_id" || (*ps)[1].Value != "configured" {
			b.Fatalf("unexpected params: %+v", *ps)
		}
		app.putParams(ps)
	}
}

func BenchmarkTreeGetValueCatchAllPooled(b *testing.B) {
	root := new(node)
	root.addRoute("/static/*filepath", func(c *Ctx) (any, error) { return nil, nil })
	app := New()
	app.maxParams = 1

	path := "/static/css/app.css"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb, ps, tsr := root.getValueFast(path, app)
		if cb == nil || ps == nil || tsr {
			b.Fatalf("unexpected getValue result: cb=%v psNil=%v tsr=%v", cb != nil, ps == nil, tsr)
		}
		if (*ps)[0].Key != "filepath" || (*ps)[0].Value != "/css/app.css" {
			b.Fatalf("unexpected param: %+v", (*ps)[0])
		}
		app.putParams(ps)
	}
}
