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

func TestAcceptMediaTypeHonorsQualityAndRejectsPrefixLookalikes(t *testing.T) {
	t.Parallel()

	if got := acceptMediaType("application/xml;q=0.2, application/json;q=0.9"); got != mediaJSON {
		t.Fatalf("expected JSON with the higher quality, got %v", got)
	}
	if got := acceptMediaType("application/json;q=0, application/xml"); got != mediaXML {
		t.Fatalf("expected XML when JSON is explicitly unacceptable, got %v", got)
	}
	if got := parseMediaType("application/jsonp"); got != mediaUnknown {
		t.Fatalf("expected JSON prefix lookalike to be rejected, got %v", got)
	}
	if got := parseMediaType("Application/JSON; Charset=UTF-8"); got != mediaJSON {
		t.Fatalf("expected media types to be case-insensitive, got %v", got)
	}
}

func TestTryParseBodyContentTypeWithParameters(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"ok":true}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()

	c := createCtx(nil, rec, req, nil)
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

func TestTryParseBodyRejectsTrailingJSONValue(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"ok":true} {"ok":false}`))
	req.Header.Set("Content-Type", "application/json")
	c := createCtx(nil, httptest.NewRecorder(), req, nil)
	defer releaseCtx(c)

	var payload struct {
		Ok bool `json:"ok"`
	}
	if err := c.TryParseBody(&payload); err == nil {
		t.Fatal("expected a trailing JSON value to be rejected")
	}
}

func TestTryParseJSONBodyFast(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"ok":true,"extra":1}`))
	rec := httptest.NewRecorder()

	c := createCtx(nil, rec, req, nil)
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

func TestApplicationCustomWriter(t *testing.T) {
	t.Parallel()

	app := New()
	if err := app.RegisterWriter("application/json", func(c *Ctx, v any) error {
		_, err := c.Write([]byte(`custom`))
		return err
	}); err != nil {
		t.Fatalf("register writer: %v", err)
	}

	app.Get("/custom", func(c *Ctx) (any, error) {
		return map[string]bool{"ok": true}, nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/custom", nil)
	req.Header.Set("Accept", "application/json")
	app.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != "custom" {
		t.Fatalf("expected custom writer body, got %q", got)
	}
}

func TestApplicationCustomReader(t *testing.T) {
	t.Parallel()

	app := New()
	if err := app.RegisterReader("application/json", func(c *Ctx, v any) error {
		payload := v.(*struct {
			Ok bool `json:"ok"`
		})
		payload.Ok = true
		return nil
	}); err != nil {
		t.Fatalf("register reader: %v", err)
	}

	app.Post("/custom", func(c *Ctx) (any, error) {
		var payload struct {
			Ok bool `json:"ok"`
		}
		if err := c.TryParseBody(&payload); err != nil {
			return nil, err
		}
		return payload, nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/custom", bytes.NewBufferString(`{"ok":false}`))
	req.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusOK {
		t.Fatalf("expected status 200, got %d", got)
	}
	if got := rec.Body.String(); got != "{\"ok\":true}\n" {
		t.Fatalf("expected custom reader to set payload, got %q", got)
	}
}
