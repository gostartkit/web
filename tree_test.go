package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type routeDef struct {
	path string
	tag  string
}

func TestApplicationStaticRoutePreferredOverParamSibling(t *testing.T) {
	app := New()
	app.Get("/organizations/:id/devices/:device_id", func(c *Ctx) (any, error) {
		return "device:" + c.Param("id") + ":" + c.Param("device_id"), nil
	})
	app.Get("/organizations/:id/devices/provision", func(c *Ctx) (any, error) {
		return "provision:" + c.Param("id"), nil
	})

	assertRouteBody(t, app, "/organizations/1/devices/provision", `"provision:1"`+"\n")
	assertRouteBody(t, app, "/organizations/1/devices/42", `"device:1:42"`+"\n")
}

func TestApplicationStaticSiblingsDoNotGetConsumedByParam(t *testing.T) {
	app := New()
	app.Get("/organizations/:id/devices/bulk/disable", func(c *Ctx) (any, error) {
		return "bulk-disable:" + c.Param("id"), nil
	})
	app.Get("/organizations/:id/devices/config/rollout", func(c *Ctx) (any, error) {
		return "config-rollout:" + c.Param("id"), nil
	})
	app.Get("/organizations/:id/devices/:device_id", func(c *Ctx) (any, error) {
		return "device:" + c.Param("id") + ":" + c.Param("device_id"), nil
	})

	assertRouteBody(t, app, "/organizations/1/devices/bulk/disable", `"bulk-disable:1"`+"\n")
	assertRouteBody(t, app, "/organizations/1/devices/config/rollout", `"config-rollout:1"`+"\n")
	assertRouteBody(t, app, "/organizations/1/devices/42", `"device:1:42"`+"\n")
}

func TestApplicationFullRouteSetCoexistsWithStaticPriority(t *testing.T) {
	app := buildDeviceRoutes(
		routeDef{path: "/organizations/:id/devices/bulk/disable", tag: "bulk-disable"},
		routeDef{path: "/organizations/:id/devices/provision", tag: "provision"},
		routeDef{path: "/organizations/:id/devices/config/rollout", tag: "config-rollout"},
		routeDef{path: "/organizations/:id/devices/:device_id", tag: "device"},
	)

	assertRouteBody(t, app, "/organizations/1/devices/bulk/disable", `"bulk-disable:1"`+"\n")
	assertRouteBody(t, app, "/organizations/1/devices/provision", `"provision:1"`+"\n")
	assertRouteBody(t, app, "/organizations/1/devices/config/rollout", `"config-rollout:1"`+"\n")
	assertRouteBody(t, app, "/organizations/1/devices/42", `"device:1:42"`+"\n")
}

func TestApplicationParamFallbackWhenStaticPrefixDoesNotFullyMatch(t *testing.T) {
	app := New()
	app.Get("/organizations/:id/devices/config/rollout", func(c *Ctx) (any, error) {
		return "config-rollout:" + c.Param("id"), nil
	})
	app.Get("/organizations/:id/devices/:device_id", func(c *Ctx) (any, error) {
		return "device:" + c.Param("id") + ":" + c.Param("device_id"), nil
	})

	assertRouteBody(t, app, "/organizations/1/devices/config/rollout", `"config-rollout:1"`+"\n")
	assertRouteBody(t, app, "/organizations/1/devices/configured", `"device:1:configured"`+"\n")
}

func TestApplicationParamFallbackFromStaticPrefixes(t *testing.T) {
	app := buildDeviceRoutes(
		routeDef{path: "/organizations/:id/devices/bulk/disable", tag: "bulk-disable"},
		routeDef{path: "/organizations/:id/devices/config/rollout", tag: "config-rollout"},
		routeDef{path: "/organizations/:id/devices/:device_id", tag: "device"},
	)

	assertRouteBody(t, app, "/organizations/1/devices/bulk", `"device:1:bulk"`+"\n")
	assertRouteBody(t, app, "/organizations/1/devices/config", `"device:1:config"`+"\n")
	assertRouteBody(t, app, "/organizations/1/devices/configured", `"device:1:configured"`+"\n")
}

func TestApplicationNestedParamFallbackFromStaticPrefixes(t *testing.T) {
	app := New()
	app.Get("/:org/x/:device/ok", func(c *Ctx) (any, error) {
		return "param:" + c.Param("org") + ":" + c.Param("device"), nil
	})
	app.Get("/foo/x/:device/nope", func(c *Ctx) (any, error) {
		return "foo-device:" + c.Param("device"), nil
	})
	app.Get("/foo/x/bar/nope", func(c *Ctx) (any, error) {
		return "foo-bar", nil
	})

	assertRouteBody(t, app, "/foo/x/bar/ok", `"param:foo:bar"`+"\n")
}

func TestApplicationStaticAndParamSiblingRegistrationOrderDoesNotMatter(t *testing.T) {
	build := func(paramFirst bool) *Application {
		app := New()
		paramRoute := func(c *Ctx) (any, error) {
			return "device:" + c.Param("id") + ":" + c.Param("device_id"), nil
		}
		staticRoute := func(c *Ctx) (any, error) {
			return "provision:" + c.Param("id"), nil
		}

		if paramFirst {
			app.Get("/organizations/:id/devices/:device_id", paramRoute)
			app.Get("/organizations/:id/devices/provision", staticRoute)
		} else {
			app.Get("/organizations/:id/devices/provision", staticRoute)
			app.Get("/organizations/:id/devices/:device_id", paramRoute)
		}

		return app
	}

	tests := []struct {
		path string
		want string
	}{
		{path: "/organizations/1/devices/provision", want: `"provision:1"` + "\n"},
		{path: "/organizations/1/devices/42", want: `"device:1:42"` + "\n"},
	}

	for _, tt := range tests {
		gotParamFirst := serveRouteBody(t, build(true), tt.path)
		gotStaticFirst := serveRouteBody(t, build(false), tt.path)
		if gotParamFirst != tt.want {
			t.Fatalf("param-first %s: got %q want %q", tt.path, gotParamFirst, tt.want)
		}
		if gotStaticFirst != tt.want {
			t.Fatalf("static-first %s: got %q want %q", tt.path, gotStaticFirst, tt.want)
		}
	}
}

func TestApplicationFullRouteSetRegistrationOrderDoesNotMatter(t *testing.T) {
	orders := []struct {
		name   string
		routes []routeDef
	}{
		{
			name: "param-first",
			routes: []routeDef{
				{path: "/organizations/:id/devices/:device_id", tag: "device"},
				{path: "/organizations/:id/devices/bulk/disable", tag: "bulk-disable"},
				{path: "/organizations/:id/devices/provision", tag: "provision"},
				{path: "/organizations/:id/devices/config/rollout", tag: "config-rollout"},
			},
		},
		{
			name: "static-first",
			routes: []routeDef{
				{path: "/organizations/:id/devices/bulk/disable", tag: "bulk-disable"},
				{path: "/organizations/:id/devices/provision", tag: "provision"},
				{path: "/organizations/:id/devices/config/rollout", tag: "config-rollout"},
				{path: "/organizations/:id/devices/:device_id", tag: "device"},
			},
		},
	}

	tests := []struct {
		path string
		want string
	}{
		{path: "/organizations/1/devices/bulk/disable", want: `"bulk-disable:1"` + "\n"},
		{path: "/organizations/1/devices/provision", want: `"provision:1"` + "\n"},
		{path: "/organizations/1/devices/config/rollout", want: `"config-rollout:1"` + "\n"},
		{path: "/organizations/1/devices/42", want: `"device:1:42"` + "\n"},
		{path: "/organizations/1/devices/bulk", want: `"device:1:bulk"` + "\n"},
		{path: "/organizations/1/devices/config", want: `"device:1:config"` + "\n"},
	}

	for _, order := range orders {
		app := buildDeviceRoutes(order.routes...)
		for _, tt := range tests {
			if got := serveRouteBody(t, app, tt.path); got != tt.want {
				t.Fatalf("%s %s: got %q want %q", order.name, tt.path, got, tt.want)
			}
		}
	}
}

func TestTreeWildcardRulesRemainStrict(t *testing.T) {
	t.Run("catch-all must stay at path end", func(t *testing.T) {
		root := new(node)
		assertPanicsWith(t, "catch-all routes are only allowed at the end of the path", func() {
			root.addRoute("/organizations/:id/devices/*device_path/more", testRoute("bad-catch-all"))
		})
	})

	t.Run("invalid wildcard segment still rejected", func(t *testing.T) {
		root := new(node)
		assertPanicsWith(t, "only one wildcard per path segment is allowed", func() {
			root.addRoute("/organizations/:id/devices/:device_id:extra", testRoute("bad-wildcard"))
		})
	})

	t.Run("param and catch-all still conflict on same layer", func(t *testing.T) {
		root := new(node)
		root.addRoute("/organizations/:id/devices/:device_id", testRoute("device"))
		assertPanicsWith(t, "conflicts with existing wildcard", func() {
			root.addRoute("/organizations/:id/devices/*device_path", testRoute("device-path"))
		})
	})
}

func serveRouteBody(t *testing.T, app *Application, path string) string {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("%s: got status %d want %d", path, rec.Code, http.StatusOK)
	}
	return rec.Body.String()
}

func assertRouteBody(t *testing.T, app *Application, path, want string) {
	t.Helper()

	if got := serveRouteBody(t, app, path); got != want {
		t.Fatalf("%s: got body %q want %q", path, got, want)
	}
}

func assertPanicsWith(t *testing.T, want string, fn func()) {
	t.Helper()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic containing %q", want)
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !strings.Contains(msg, want) {
			t.Fatalf("panic %q does not contain %q", msg, want)
		}
	}()

	fn()
}

func testRoute(name string) Next {
	return func(c *Ctx) (any, error) {
		return name, nil
	}
}

func buildDeviceRoutes(routes ...routeDef) *Application {
	app := New()
	for _, route := range routes {
		app.Get(route.path, func(tag string) Next {
			return func(c *Ctx) (any, error) {
				switch tag {
				case "device":
					return "device:" + c.Param("id") + ":" + c.Param("device_id"), nil
				default:
					return tag + ":" + c.Param("id"), nil
				}
			}
		}(route.tag))
	}
	return app
}
