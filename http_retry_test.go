package web

import (
	"bytes"
	"context"
	"encoding/json"
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func TestDoReqWithClientUsesProvidedClient(t *testing.T) {
	t.Parallel()

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
		t.Fatalf("new request failed: %v", err)
	}

	var out struct {
		Ok bool `json:"ok"`
	}
	if err := DoReqWithClient(client, req, &out, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !out.Ok {
		t.Fatalf("expected parsed response payload")
	}
}

func TestDoReqWithClientRawBytesFastPath(t *testing.T) {
	t.Parallel()

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
		t.Fatalf("new request failed: %v", err)
	}

	var out []byte
	if err := DoReqWithClient(client, req, &out, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != `{"ok":true}` {
		t.Fatalf("expected raw response bytes, got %q", string(out))
	}
}

func TestDoReqWithClientRawMessageFastPath(t *testing.T) {
	t.Parallel()

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
		t.Fatalf("new request failed: %v", err)
	}

	var out json.RawMessage
	if err := DoReqWithClient(client, req, &out, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != `{"ok":true}` {
		t.Fatalf("expected raw json message, got %q", string(out))
	}
}

func TestDoReqWithClientBufferFastPath(t *testing.T) {
	t.Parallel()

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
		t.Fatalf("new request failed: %v", err)
	}

	var out bytes.Buffer
	if err := DoReqWithClient(client, req, &out, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.String() != `{"ok":true}` {
		t.Fatalf("expected raw buffer body, got %q", out.String())
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

func TestDoBytesSetsDefaultHeaders(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/octet-stream" {
			t.Fatalf("expected Content-Type application/octet-stream, got %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("expected Accept application/json, got %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	if err := DoBytes(context.Background(), http.MethodPost, srv.URL, "", []byte("abc"), nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTryDoBytesReusesBodyForRetry(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body failed: %v", err)
		}
		if string(body) != "payload" {
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

	if err := TryDoBytes(context.Background(), http.MethodPost, srv.URL, "", []byte("payload"), &out, 2); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts.Load())
	}
	if !out.Ok {
		t.Fatalf("expected parsed response payload")
	}
}

func TestTryDoBytesWithClientUsesProvidedClient(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body failed: %v", err)
			}
			if !bytes.Equal(body, []byte("payload")) {
				t.Fatalf("expected request body payload, got %q", string(body))
			}

			if attempts.Add(1) == 1 {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
					Request:    r,
				}, nil
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		}),
	}

	var out struct {
		Ok bool `json:"ok"`
	}
	if err := TryDoBytesWithClient(client, context.Background(), http.MethodPost, "http://example.com", "", []byte("payload"), &out, 2); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts.Load())
	}
	if !out.Ok {
		t.Fatalf("expected parsed response payload")
	}
}
