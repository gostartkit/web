package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	app = CreateApplication()
)

func TestHttpGet(t *testing.T) {

	rel := "/route/"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, rel, nil)

	handler := func(c *Ctx) (any, error) {

		if c.Method() != http.MethodGet {
			t.Errorf("Expected GET route to be added, but got %s", c.Method())
		}

		if c.r.URL.Path != rel {
			t.Errorf("Expected path %s, but got %s", rel, c.r.URL.Path)
		}

		return "", nil
	}

	app.Get(rel, handler)

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", rec.Code)
	}
}

func TestHttpPost(t *testing.T) {

	rel := "/route/"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, rel, nil)

	handler := func(c *Ctx) (any, error) {

		if c.Method() != http.MethodPost {
			t.Errorf("Expected POST route to be added, but got %s", c.Method())
		}

		if c.r.URL.Path != rel {
			t.Errorf("Expected path /route/, but got %s", c.r.URL.Path)
		}

		return 1, nil
	}

	app.Post(rel, handler)

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status code 201, but got %d", rec.Code)
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
	url := "http://gostartkit.com"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, rel, nil)

	handler := func(c *Ctx) (any, error) { return url, ErrMovedPermanently }
	app.Get(rel, handler)

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status code 301, but got %d", rec.Code)
	}
	if rec.Header().Get("Location") != url {
		t.Errorf("Expected Location header '%s', but got %s", url, rec.Header().Get("Location"))
	}
}
