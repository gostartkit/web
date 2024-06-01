package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	app = CreateApplication()
)

func TestAddRoutes(t *testing.T) {

	rel := "/route/"

	getHandler := func(c *Ctx) (any, error) {
		if c.Method() == http.MethodGet {
			t.Errorf("Expected GET route to be added")
		}

		if c.r.URL.Path != rel {
			t.Errorf("Expected path %s, but got %s", rel, c.r.URL.Path)
		}

		return nil, nil
	}

	postHandler := func(c *Ctx) (any, error) {
		if c.Method() == http.MethodGet {
			t.Errorf("Expected POST route to be added")
		}

		if c.r.URL.Path != rel {
			t.Errorf("Expected path /route/, but got %s", c.r.URL.Path)
		}
		return nil, nil
	}

	app.Get(rel, getHandler)
	app.Post(rel, postHandler)
}

func TestServeHTTP(t *testing.T) {

	rel := "/test/"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, rel, nil)

	handler := func(c *Ctx) (any, error) { return "test", nil }
	app.Get(rel, handler)

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", rec.Code)
	}

	data := strings.TrimRight(rec.Body.String(), "\n")

	if data != "\"test\"" {
		t.Errorf("Expected body \"test\", but got %s", data)
	}
}

func TestErrorHandling(t *testing.T) {

	rel := "/error/"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, rel, nil)

	handler := func(c *Ctx) (any, error) { return nil, ErrNotFound }
	app.Get(rel, handler)

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status code 404, but got %d", rec.Code)
	}
}

func TestRedirectHandling(t *testing.T) {

	rel := "/redirect/"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, rel, nil)

	handler := func(c *Ctx) (any, error) { return "http://gostartkit.com", ErrMovedPermanently }
	app.Get(rel, handler)

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status code 301, but got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "http://gostartkit.com" {
		t.Errorf("Expected Location header 'http://gostartkit.com', but got %s", rec.Header().Get("Location"))
	}
}
