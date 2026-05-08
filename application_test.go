package web

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Expected Content-Type application/json, but got %q", got)
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

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", rec.Code)
	}
}

func TestHttpPostExplicitStatusOverride(t *testing.T) {
	app := New()

	rel := "/route/"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, rel, nil)

	app.Post(rel, func(c *Ctx) (any, error) {
		c.SetStatus(http.StatusCreated)
		return 1, nil
	})

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status code 201, but got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Expected Content-Type application/json, but got %q", got)
	}
}

func TestCustomHTTPMethod(t *testing.T) {
	app := New()

	app.Handle("PURGE", "/cache/:key", func(c *Ctx) (any, error) {
		return c.Param("key"), nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PURGE", "/cache/home", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, but got %d", rec.Code)
	}
	if got := rec.Body.String(); got != `"home"`+"\n" {
		t.Fatalf("Expected body %q, got %q", `"home"`+"\n", got)
	}
}

func TestApplicationOptionsAndHTTPHandler(t *testing.T) {
	middlewareCalled := false
	app := New(
		WithMiddleware(func(next Next) Next {
			return func(c *Ctx) (any, error) {
				middlewareCalled = true
				return next(c)
			}
		}),
		WithNotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "custom not found", http.StatusNotFound)
		})),
		WithMethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "custom method", http.StatusMethodNotAllowed)
		})),
	)

	app.GetHTTP("/std/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Handler", "std")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(r.URL.Path))
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/std/42", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("Expected status code 202, but got %d", rec.Code)
	}
	if got := rec.Header().Get("X-Handler"); got != "std" {
		t.Fatalf("Expected standard handler header, got %q", got)
	}
	if got := rec.Body.String(); got != "/std/42" {
		t.Fatalf("Expected standard handler body, got %q", got)
	}
	if !middlewareCalled {
		t.Fatalf("Expected option middleware to run")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/std/42", nil)
	app.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected method-not-allowed handler, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/missing", nil)
	app.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("Expected not-found handler, got %d", rec.Code)
	}
}

func TestRouteGroupHTTPHandler(t *testing.T) {
	app := New()
	api := app.Group("/api")

	api.GetHTTP("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, but got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "pong" {
		t.Fatalf("Expected body pong, got %q", got)
	}
}

func TestManualWriteDefaultsToOK(t *testing.T) {
	app := New()

	app.Get("/manual", func(c *Ctx) (any, error) {
		_, err := c.Write([]byte("ok"))
		return nil, err
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/manual", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, but got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Fatalf("Expected body ok, but got %q", got)
	}
}

func TestManualWriteUsesExplicitStatus(t *testing.T) {
	app := New()

	app.Get("/manual", func(c *Ctx) (any, error) {
		c.SetStatus(http.StatusAccepted)
		_, err := c.Write([]byte("ok"))
		return nil, err
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/manual", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("Expected status code 202, but got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Fatalf("Expected body ok, but got %q", got)
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

func TestHttpPathParamsAcrossRequests(t *testing.T) {
	app := New()

	app.Get("/users/:userID/books/:bookID", func(c *Ctx) (any, error) {
		userID, err := c.ParamUint64("userID")
		if err != nil {
			return nil, err
		}
		bookID, err := c.ParamUint64("bookID")
		if err != nil {
			return nil, err
		}
		return map[string]uint64{
			"userID": userID,
			"bookID": bookID,
		}, nil
	})

	tests := []struct {
		path string
		body string
	}{
		{path: "/users/42/books/100", body: `{"bookID":100,"userID":42}` + "\n"},
		{path: "/users/7/books/9", body: `{"bookID":9,"userID":7}` + "\n"},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		app.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status code 200, got %d", tt.path, rec.Code)
		}
		if got := rec.Body.String(); got != tt.body {
			t.Fatalf("%s: expected body %q, got %q", tt.path, tt.body, got)
		}
	}
}

func TestHttpPathParamsMoreThanThree(t *testing.T) {
	app := New()

	app.Get("/a/:p1/b/:p2/c/:p3/d/:p4", func(c *Ctx) (any, error) {
		return c.Param("p1") + "," + c.Param("p2") + "," + c.Param("p3") + "," + c.Param("p4"), nil
	})

	tests := []struct {
		path string
		body string
	}{
		{path: "/a/one/b/two/c/three/d/four", body: `"one,two,three,four"` + "\n"},
		{path: "/a/red/b/blue/c/green/d/gold", body: `"red,blue,green,gold"` + "\n"},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		app.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status code 200, got %d", tt.path, rec.Code)
		}
		if got := rec.Body.String(); got != tt.body {
			t.Fatalf("%s: expected body %q, got %q", tt.path, tt.body, got)
		}
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

func TestCtxRedirectHandling(t *testing.T) {
	app := New()

	rel := "/redirect-ctx/"
	url := "/new-location-ctx/"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, rel, nil)

	app.Get(rel, func(c *Ctx) (any, error) {
		return nil, c.Redirect(http.StatusFound, url)
	})

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("Expected status code 302, but got %d", rec.Code)
	}
	if rec.Header().Get("Location") != url {
		t.Errorf("Expected Location header '%s', but got %s", url, rec.Header().Get("Location"))
	}
}

func TestDeprecatedRedirectAliasBypassesCustomErrorHandler(t *testing.T) {
	app := New()
	app.SetErrorHandler(JSONErrorHandler(true))

	rel := "/redirect-legacy/"
	url := "/new-location-legacy/"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, rel, nil)

	app.Get(rel, func(c *Ctx) (any, error) {
		return Redirect(url, http.StatusMovedPermanently)
	})

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status code 301, but got %d", rec.Code)
	}
	if rec.Header().Get("Location") != url {
		t.Errorf("Expected Location header '%s', but got %s", url, rec.Header().Get("Location"))
	}
	if got := rec.Header().Get("Content-Type"); got == "application/json" {
		t.Errorf("Expected deprecated redirect alias to bypass JSON error handler, got content type %q", got)
	}
}

func TestApplicationInfoLoggerUsesLogfmt(t *testing.T) {
	var buf bytes.Buffer

	app := New()
	app.SetInfoLogger(log.New(&buf, "", 0))
	app.Get("/log", func(c *Ctx) (any, error) {
		c.Init(42)
		return map[string]bool{"ok": true}, nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/log", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Host = "api.example.test"
	app.ServeHTTP(rec, req)

	got := strings.TrimSpace(buf.String())
	want := `level="info" event="request" remote_addr="127.0.0.1:1234" host="api.example.test" user_id=42 method="GET" path="/log" status=200`
	if got != want {
		t.Fatalf("unexpected info log:\n got: %s\nwant: %s", got, want)
	}
}

func TestApplicationErrLoggerUsesLogfmt(t *testing.T) {
	var buf bytes.Buffer

	app := New()
	app.SetErrLogger(log.New(&buf, "", 0))
	app.Get("/err-log", func(c *Ctx) (any, error) {
		c.Init(7)
		return nil, ErrNotFound
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/err-log", nil)
	req.RemoteAddr = "127.0.0.1:5678"
	req.Host = "api.example.test"
	app.ServeHTTP(rec, req)

	got := strings.TrimSpace(buf.String())
	want := `level="error" event="request" remote_addr="127.0.0.1:5678" host="api.example.test" user_id=7 method="GET" path="/err-log" status=404 error="NOTFOUND"`
	if got != want {
		t.Fatalf("unexpected error log:\n got: %s\nwant: %s", got, want)
	}
}

func TestApplicationPanicLoggerUsesLogfmt(t *testing.T) {
	var buf bytes.Buffer

	app := New()
	app.SetErrLogger(log.New(&buf, "", 0))
	app.Get("/panic-log", func(c *Ctx) (any, error) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic-log", nil)
	req.RemoteAddr = "127.0.0.1:9999"
	req.Host = "api.example.test"
	app.ServeHTTP(rec, req)

	got := strings.TrimSpace(buf.String())
	want := `level="error" event="panic" remote_addr="127.0.0.1:9999" host="api.example.test" method="GET" path="/panic-log" error="boom"`
	if got != want {
		t.Fatalf("unexpected panic log:\n got: %s\nwant: %s", got, want)
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
