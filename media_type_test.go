package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteCodeAcceptWithParameters(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/xml; charset=utf-8")

	writeCode(rec, req, http.StatusOK)

	if got := rec.Header().Get("Content-Type"); got != "application/xml" {
		t.Fatalf("expected application/xml, got %q", got)
	}
}

func TestTryParseBodyContentTypeWithParameters(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"ok":true}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()

	c := createCtx(rec, req, nil)
	defer releaseCtx(c)

	var payload struct {
		Ok bool `json:"ok"`
	}
	if err := c.TryParseBody(&payload); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !payload.Ok {
		t.Fatalf("expected parsed payload")
	}
}

func TestTryParseJSONBodyFast(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"ok":true,"extra":1}`))
	rec := httptest.NewRecorder()

	c := createCtx(rec, req, nil)
	defer releaseCtx(c)

	var payload struct {
		Ok bool `json:"ok"`
	}
	if err := c.TryParseJSONBodyFast(&payload); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !payload.Ok {
		t.Fatalf("expected parsed payload")
	}
}
