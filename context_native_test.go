package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (r *flushRecorder) Flush() {
	r.flushed = true
}

func TestCtxNativeRequestResponseWriter(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/native", nil)
	rec := httptest.NewRecorder()
	c := createCtx(nil, rec, req, nil)
	defer releaseCtx(c)

	if c.Request() != req {
		t.Fatalf("expected request to match input request")
	}

	if c.ResponseWriter() != rec {
		t.Fatalf("expected response writer to match input writer")
	}
}

func TestCtxImplementsResponseWriterMethods(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/rw", nil)
	rec := httptest.NewRecorder()
	c := createCtx(nil, rec, req, nil)
	defer releaseCtx(c)

	c.Header().Set("X-Test", "1")
	c.WriteHeader(http.StatusAccepted)
	_, err := c.Write([]byte("ok"))
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if got := rec.Code; got != http.StatusAccepted {
		t.Fatalf("expected code %d, got %d", http.StatusAccepted, got)
	}
	if got := rec.Header().Get("X-Test"); got != "1" {
		t.Fatalf("expected header X-Test=1, got %q", got)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Fatalf("expected body ok, got %q", got)
	}
}

func TestCtxFlushAndUnsupportedMethods(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/flush", nil)
	rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
	c := createCtx(nil, rec, req, nil)
	defer releaseCtx(c)

	c.Flush()
	if !rec.flushed {
		t.Fatalf("expected flush to be forwarded")
	}

	if _, _, err := c.Hijack(); err != http.ErrNotSupported {
		t.Fatalf("expected Hijack not supported, got %v", err)
	}

	if err := c.Push("/asset.js", nil); err != http.ErrNotSupported {
		t.Fatalf("expected Push not supported, got %v", err)
	}
}
