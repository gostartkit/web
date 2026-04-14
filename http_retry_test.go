package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDoSetsDefaultHeadersWhenBeforeIsEmptySlice(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected Content-Type application/json, got %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("expected Accept application/json, got %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	before := []func(*http.Request){}
	if err := Do(context.Background(), http.MethodGet, srv.URL, "", nil, nil, before...); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTryGetRetryZeroStillAttemptsOnce(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	var out struct {
		Ok bool `json:"ok"`
	}

	if err := TryGet(context.Background(), srv.URL, "", &out, 0); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !out.Ok {
		t.Fatalf("expected parsed response payload")
	}
}

func TestTryDoRecreatesBodyForRetry(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body failed: %v", err)
		}
		if string(body) != `{"name":"sam"}` {
			t.Fatalf("expected stable request body, got %q", string(body))
		}

		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	var out struct {
		Ok bool `json:"ok"`
	}

	if err := TryDo(context.Background(), http.MethodPost, srv.URL, "", strings.NewReader(`{"name":"sam"}`), &out, 2); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts.Load())
	}
	if !out.Ok {
		t.Fatalf("expected parsed response payload")
	}
}
