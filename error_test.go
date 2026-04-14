package web

import (
	"fmt"
	"net/http"
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
