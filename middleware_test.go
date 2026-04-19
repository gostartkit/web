package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestRequestIDMiddlewareGeneratesAndPropagatesID(t *testing.T) {
	t.Parallel()

	app := New()
	app.Use(RequestID("", func() string { return "req-1" }))
	app.Get("/id", func(c *Ctx) (any, error) {
		return map[string]string{"id": c.RequestID()}, nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/id", nil)
	app.ServeHTTP(rec, req)

	if got := rec.Header().Get(DefaultRequestIDHeader); got != "req-1" {
		t.Fatalf("expected response request id %q, got %q", "req-1", got)
	}
	if got := rec.Body.String(); got != "{\"id\":\"req-1\"}\n" {
		t.Fatalf("expected request id in response body, got %q", got)
	}
}

func TestRequestIDMiddlewarePreservesIncomingHeader(t *testing.T) {
	t.Parallel()

	app := New()
	app.Use(RequestID("", func() string { return "generated" }))
	app.Get("/id", func(c *Ctx) (any, error) {
		return map[string]string{"id": c.RequestID()}, nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/id", nil)
	req.Header.Set(DefaultRequestIDHeader, "external-id")
	app.ServeHTTP(rec, req)

	if got := rec.Header().Get(DefaultRequestIDHeader); got != "external-id" {
		t.Fatalf("expected preserved request id %q, got %q", "external-id", got)
	}
}

func TestRecoverMiddlewareConvertsPanicToFrameworkResponse(t *testing.T) {
	t.Parallel()

	app := New()
	app.Use(Recover(nil))
	app.Get("/panic", func(c *Ctx) (any, error) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "\"INTERNALSERVERERROR\"\n" {
		t.Fatalf("expected default recover body, got %q", got)
	}
}

func TestRecoverMiddlewareAllowsCustomHandler(t *testing.T) {
	t.Parallel()

	app := New()
	app.Use(Recover(func(c *Ctx, recovered any) error {
		c.SetHeader("Content-Type", "text/plain")
		c.WriteHeader(http.StatusTeapot)
		_, err := c.Write([]byte("panic handled"))
		return err
	}))
	app.Get("/panic", func(c *Ctx) (any, error) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected status 418, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "panic handled" {
		t.Fatalf("expected custom recover body, got %q", got)
	}
}

func TestRecoverWithOptionsUsesCustomDefaults(t *testing.T) {
	t.Parallel()

	app := New()
	app.Use(RecoverWithOptions(RecoverOptions{
		DefaultStatus: http.StatusServiceUnavailable,
		DefaultBody:   "UNAVAILABLE",
	}))
	app.Get("/panic", func(c *Ctx) (any, error) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "\"UNAVAILABLE\"\n" {
		t.Fatalf("expected custom default recover body, got %q", got)
	}
}

func TestTimeoutMiddlewareReturnsRequestTimeout(t *testing.T) {
	t.Parallel()

	app := New()
	app.Use(Timeout(5 * time.Millisecond))
	app.Get("/slow", func(c *Ctx) (any, error) {
		<-c.Context().Done()
		return nil, c.Context().Err()
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestTimeout {
		t.Fatalf("expected status 408, got %d", rec.Code)
	}
}

func TestAccessLogMiddlewareReportsFrameworkStatus(t *testing.T) {
	t.Parallel()

	app := New()
	got := make([]int, 0, 3)
	app.Use(AccessLog(func(c *Ctx, status int, d time.Duration, err error) {
		got = append(got, status)
	}))

	app.Get("/empty", func(c *Ctx) (any, error) { return nil, nil })
	app.Post("/created", func(c *Ctx) (any, error) { return "ok", nil })
	app.Get("/err", func(c *Ctx) (any, error) { return nil, ErrNotFound })

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/empty"},
		{method: http.MethodPost, path: "/created"},
		{method: http.MethodGet, path: "/err"},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(tt.method, tt.path, nil)
		app.ServeHTTP(rec, req)
	}

	want := []int{http.StatusNoContent, http.StatusOK, http.StatusNotFound}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected logged statuses: got %v want %v", got, want)
	}
}

func TestAccessLogMiddlewareUsesExplicitStatusOverride(t *testing.T) {
	t.Parallel()

	app := New()
	got := make([]int, 0, 1)
	app.Use(AccessLog(func(c *Ctx, status int, d time.Duration, err error) {
		got = append(got, status)
	}))

	app.Post("/accepted", func(c *Ctx) (any, error) {
		c.SetStatus(http.StatusAccepted)
		return "ok", nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/accepted", nil)
	app.ServeHTTP(rec, req)

	want := []int{http.StatusAccepted}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected logged statuses: got %v want %v", got, want)
	}
}

func TestAccessLogWithOptionsUsesCustomStatusMapperAndClock(t *testing.T) {
	t.Parallel()

	app := New()
	var entry AccessLogEntry
	nowCalls := 0
	base := time.Unix(10, 0)
	app.Use(AccessLogWithOptions(AccessLogOptions{
		Now: func() time.Time {
			nowCalls++
			if nowCalls == 1 {
				return base
			}
			return base.Add(25 * time.Millisecond)
		},
		StatusMapper: func(c *Ctx, val any, err error) int {
			return http.StatusAccepted
		},
		Log: func(c *Ctx, e AccessLogEntry) {
			entry = e
		},
	}))

	app.Get("/ok", func(c *Ctx) (any, error) { return "ok", nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	app.ServeHTTP(rec, req)

	if entry.Status != http.StatusAccepted {
		t.Fatalf("expected custom status 202, got %d", entry.Status)
	}
	if entry.Duration != 25*time.Millisecond {
		t.Fatalf("expected custom duration 25ms, got %s", entry.Duration)
	}
}

func TestTimeoutMiddlewarePassesThroughNonTimeoutError(t *testing.T) {
	t.Parallel()

	app := New()
	app.Use(Timeout(10 * time.Millisecond))
	app.Get("/bad", func(c *Ctx) (any, error) {
		return nil, errors.New("boom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for generic error, got %d", rec.Code)
	}
}

func TestJSONErrorHandlerWritesStructuredBody(t *testing.T) {
	t.Parallel()

	app := New()
	app.SetErrorHandler(JSONErrorHandler(false))
	app.Get("/err", func(c *Ctx) (any, error) {
		return nil, ErrNotFound
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "{\"code\":404,\"message\":\"NOTFOUND\"}\n" {
		t.Fatalf("unexpected structured error body: %q", got)
	}
}

func TestJSONErrorHandlerIncludesRequestID(t *testing.T) {
	t.Parallel()

	app := New()
	app.Use(RequestID("", func() string { return "req-42" }))
	app.SetErrorHandler(JSONErrorHandler(true))
	app.Get("/err", func(c *Ctx) (any, error) {
		return nil, ErrForbidden
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "{\"code\":403,\"message\":\"FORBIDDEN\",\"request_id\":\"req-42\"}\n" {
		t.Fatalf("unexpected structured error body: %q", got)
	}
}
