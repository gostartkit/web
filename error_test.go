package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrCodeSupportsWrappedErrors(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("wrapped: %w", NewErr(http.StatusConflict, "CONFLICT"))
	if code := errCode(err); code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, code)
	}
}

func TestErrCodeSupportsWrappedErrFn(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("wrapped: %w", NewErrFn(http.StatusTooManyRequests, "TOOMANYREQUESTS", nil))
	if code := errCode(err); code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, code)
	}
}

func TestApplicationHandlesWrappedErrFn(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/wrapped", func(c *Ctx) (any, error) {
		return nil, fmt.Errorf("wrapped: %w", NewErrFn(http.StatusConflict, "CONFLICT", func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusConflict)
			_, err := w.Write([]byte("conflict"))
			return err
		}))
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/wrapped", nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
	if got := rec.Body.String(); got != "conflict" {
		t.Fatalf("expected custom error body, got %q", got)
	}
}
