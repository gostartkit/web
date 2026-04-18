package web

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHttpGet(t *testing.T) {
	app := New()

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
	app := New()

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

func TestHttpPathParamWithoutServe(t *testing.T) {
	app := New()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/42", nil)

	app.Get("/user/:id", func(c *Ctx) (any, error) {
		if got := c.Param("id"); got != "42" {
			t.Fatalf("expected param id=42, got %q", got)
		}
		return "ok", nil
	})

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, but got %d", rec.Code)
	}
}

func TestErrorHandling(t *testing.T) {
	app := New()

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
	app := New()

	rel := "/redirect/"
	url := "/new-location/"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, rel, nil)

	handler := func(c *Ctx) (any, error) { return Redirect(url, http.StatusMovedPermanently) }
	app.Get(rel, handler)

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status code 301, but got %d", rec.Code)
	}
	if rec.Header().Get("Location") != url {
		t.Errorf("Expected Location header '%s', but got %s", url, rec.Header().Get("Location"))
	}
}

func TestApplicationMiddlewareAndGroupOrder(t *testing.T) {
	app := New()
	order := make([]string, 0, 7)

	app.Use(func(next Next) Next {
		return func(c *Ctx) (any, error) {
			order = append(order, "app:before")
			val, err := next(c)
			order = append(order, "app:after")
			return val, err
		}
	})

	api := app.Group("/api")
	api.Use(func(next Next) Next {
		return func(c *Ctx) (any, error) {
			order = append(order, "group:before")
			val, err := next(c)
			order = append(order, "group:after")
			return val, err
		}
	})

	api.Handle(http.MethodGet, "/users/:id", func(c *Ctx) (any, error) {
		order = append(order, "handler:"+c.Param("id"))
		return "ok", nil
	}, func(next Next) Next {
		return func(c *Ctx) (any, error) {
			order = append(order, "route:before")
			val, err := next(c)
			order = append(order, "route:after")
			return val, err
		}
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/users/42", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, but got %d", rec.Code)
	}

	want := []string{
		"app:before",
		"group:before",
		"route:before",
		"handler:42",
		"route:after",
		"group:after",
		"app:after",
	}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("unexpected middleware order: got %v want %v", order, want)
	}
}

func TestCustomErrorHandler(t *testing.T) {
	app := New()
	app.SetErrorHandler(func(c *Ctx, err error) error {
		c.SetHeader("Content-Type", "text/plain")
		c.WriteHeader(http.StatusTeapot)
		_, writeErr := c.Write([]byte("teapot"))
		return writeErr
	})

	app.Get("/brew", func(c *Ctx) (any, error) {
		return nil, ErrNotFound
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/brew", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("Expected status code 418, but got %d", rec.Code)
	}
	if rec.Body.String() != "teapot" {
		t.Fatalf("Expected body teapot, but got %q", rec.Body.String())
	}
}

func TestMethodNotAllowed(t *testing.T) {
	app := New()
	app.Get("/users", func(c *Ctx) (any, error) {
		return "ok", nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected status code 405, but got %d", rec.Code)
	}
	if got := rec.Header().Get("Allow"); got != "GET, OPTIONS" {
		t.Fatalf("Expected Allow header %q, but got %q", "GET, OPTIONS", got)
	}
}
