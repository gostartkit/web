package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteBinary(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/bin", func(c *Ctx) (any, error) {
		return []byte{0x01, 0x02, 0x03}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/bin", nil)
	req.Header.Set("Accept", "application/octet-stream")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/octet-stream" {
		t.Fatalf("expected octet-stream content type, got %q", got)
	}
	if got := rec.Body.Bytes(); len(got) != 3 || got[0] != 0x01 || got[1] != 0x02 || got[2] != 0x03 {
		t.Fatalf("unexpected binary body: %v", got)
	}
}

type avroPayload struct {
	raw []byte
}

func (p avroPayload) MarshalAvro() ([]byte, error) {
	return p.raw, nil
}

func TestWriteAvro(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/avro", func(c *Ctx) (any, error) {
		return avroPayload{raw: []byte{0xAA, 0xBB}}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/avro", nil)
	req.Header.Set("Accept", "application/x-avro")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/x-avro" {
		t.Fatalf("expected avro content type, got %q", got)
	}
	if got := rec.Body.Bytes(); len(got) != 2 || got[0] != 0xAA || got[1] != 0xBB {
		t.Fatalf("unexpected avro body: %v", got)
	}
}

func TestWriteJSONRawMessage(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/json", func(c *Ctx) (any, error) {
		return json.RawMessage(`{"ok":true}`), nil
	})

	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected json content type, got %q", got)
	}
	if got := rec.Body.String(); got != `{"ok":true}` {
		t.Fatalf("unexpected json body: %q", got)
	}
}
