# Web.go The library for web

中文文档: [README_CN.md](./README_CN.md)

### Quick Start

```go
package main

import (
	"log"
	"net/http"

	"pkg.gostartkit.com/web"
)

func main() {
	app := web.New()

	app.Get("/health", func(c *web.Ctx) (any, error) {
		return map[string]string{"status": "ok"}, nil
	})

	log.Fatal(app.ListenAndServe("tcp", ":8080"))
}
```

### API Index

- `web.New() *Application`
- route registration:
  - `Get`, `Post`, `Put`, `Patch`, `Delete`, `Head`, `Options`
- server lifecycle:
  - `ListenAndServe`, `ListenAndServeTLS`, `Shutdown`
- helpers:
  - `ServeFiles`, `Redirect`, `TryParse(...)`, `TryXxx(...)`
- context (`*Ctx`) common methods:
  - request: `Method`, `Path`, `Query`, `Param`, `Body`, `ContentType`, `BearerToken`
  - parse: `TryParseBody`, `TryParseParam`, `TryParseQuery`, `TryParseForm`
  - response: `SetHeader`, `SetCookie`, `AllowCredentials`, content negotiation via `Accept`

### API Quick Reference (EN)

| Area | API | Description |
|---|---|---|
| Application | `New()` | Create app instance |
| Application | `Get/Post/Put/Patch/Delete/Head/Options(path, handler)` | Register route handler |
| Application | `ServeFiles("/static/*filepath", fs)` | Serve static files with catch-all path |
| Application | `ListenAndServe(network, addr, ...opts)` | Start HTTP server |
| Application | `ListenAndServeTLS(network, addr, tlsConfig, ...opts)` | Start HTTPS server |
| Application | `Shutdown(ctx)` | Graceful shutdown |
| Context | `Param(name)`, `Query(name)`, `Form(name)` | Read path/query/form values |
| Context | `TryParseBody(v)` | Parse request body by content type (JSON/GOB/XML) |
| Context | `TryParseParam/Query/Form(name, &v)` | Parse string values into typed value |
| Context | `SetHeader`, `SetCookie`, `SetContentType` | Write response headers |
| Context | `Request()`, `ResponseWriter()`, `Context()` | Access raw HTTP objects |
| Client | `Get/Post/Put/Patch/Delete/Do` | HTTP client helpers |
| Client | `TryGet/TryPost/TryPut/TryPatch/TryDelete/TryDo` | HTTP helpers with retry loop |
| Error | `NewErr(code, msg)` | Error with HTTP status code |
| Error | `Redirect(url, code)` | Return redirect response from handler |

### Response Behavior

- Handler return value controls response:
  - `(nil, nil)` -> `204 No Content`
  - `(value, nil)` -> `200 OK` (`POST` uses `201 Created`)
  - `(_, err)` -> status code from framework error type, body contains `err.Error()`
- Response format is selected by request `Accept` header:
  - `application/json`
  - `application/x-gob`
  - `application/xml`
  - `application/octet-stream`
  - `application/x-avro`

### Compatibility / Breaking Changes

- `Try*` retry semantics updated:
  - `retry <= 0` now still performs one request attempt.
  - retry loop stops early for `ErrUnauthorized`, `ErrForbidden`, and `ErrBadRequest` (including wrapped).
- `TryDo` now supports safe body replay across retries (request body is buffered once and recreated per attempt).
- `Ctx.writeBinary` and `Ctx.writeAvro` are implemented:
  - previous behavior for these media types was `ErrNotImplemented`.
  - now they support fast-path direct writing (see Binary / Avro response section).
- Redirect usage:
  - returning only `ErrMovedPermanently` does not set `Location`.
  - use `web.Redirect(url, code)` to generate proper redirect response headers.
- Header negotiation improvement:
  - `Accept`/`Content-Type` values with parameters (e.g. `application/json; charset=utf-8`) are now parsed correctly.

Migration tips:

- If you relied on `retry=0` to skip outbound call, replace with explicit conditional in caller.
- If your handlers used `application/octet-stream` or `application/x-avro`, you can now return `[]byte`, `io.Reader`, or custom marshaler types directly.
- For redirects, migrate to `web.Redirect(...)` for predictable behavior.

### Current capabilities (2026-04)

- Routing:
  - static path, `:param`, `*catchAll`
  - high-performance tree matcher (inspired by `httprouter`)
- Response encoding by `Accept`:
  - `application/json`
  - `application/x-gob`
  - `application/xml`
  - `application/octet-stream` (implemented)
  - `application/x-avro` (implemented)
- Request body parsing by `Content-Type`:
  - `application/json`
  - `application/x-gob`
  - `application/xml`

### Binary / Avro response

`Ctx.writeBinary` and `Ctx.writeAvro` are optimized for fast paths.

- Binary fast-path input types:
  - `[]byte`
  - `string`
  - `*bytes.Buffer`
  - `io.Reader`
  - `encoding.BinaryMarshaler`
- Avro fast-path input types:
  - `web.AvroMarshaler`
  - falls back to binary writer for the same input types above

```go
type Event struct {
	Raw []byte
}

func (e Event) MarshalAvro() ([]byte, error) {
	return e.Raw, nil
}

app.Get("/payload", func(c *web.Ctx) (any, error) {
	// Client sends: Accept: application/x-avro
	return Event{Raw: []byte{0xAA, 0xBB}}, nil
})
```

### Redirect helper

Use `web.Redirect(url, code)` to return redirect responses.

```go
app.Get("/old", func(c *web.Ctx) (any, error) {
	return web.Redirect("/new", http.StatusMovedPermanently)
})
```

### HTTP client retry behavior

`TryGet`, `TryPost`, `TryPut`, `TryPatch`, `TryDelete`, `TryDo`:

- `retry <= 0` still performs at least **one** request.
- retries stop early for non-retriable errors:
  - `ErrUnauthorized`
  - `ErrForbidden`
  - `ErrBadRequest` (including wrapped)
- `TryDo` safely retries with request body replay (body is cached once and recreated per attempt).

### Benchmark

Run focused benchmarks:

```bash
go test -run '^$' -bench 'Benchmark(ServeHTTP|TreeGetValue|TryParseBody|PostJSON)' -benchmem ./...
```

### Full Example: Route
```go
package route

import (
	"app.gostartkit.com/go/auth/config"
	"app.gostartkit.com/go/auth/controller"
	"app.gostartkit.com/go/auth/middleware"
	"pkg.gostartkit.com/web"
)

func userRoute(app *web.Application, prefix string) {

	c := controller.CreateUserController()

	app.Get(prefix+"/user/", middleware.Chain(c.Index, config.Read|config.ReadUser))
	app.Get(prefix+"/user/:id", middleware.Chain(c.Detail, config.Read|config.ReadUser))
	app.Post(prefix+"/apply/user/id/", middleware.Chain(c.CreateID, config.Write|config.WriteUser))
	app.Post(prefix+"/user/", middleware.Chain(c.Create, config.Write|config.WriteUser))
	app.Put(prefix+"/user/:id", middleware.Chain(c.Update, config.Write|config.WriteUser))
	app.Patch(prefix+"/user/:id", middleware.Chain(c.Patch, config.Write|config.WriteUser))
	app.Patch(prefix+"/user/:id/status/", middleware.Chain(c.UpdateStatus, config.Write|config.WriteUser))
	app.Delete(prefix+"/user/:id", middleware.Chain(c.Destroy, config.Write|config.WriteUser))
	app.Get(prefix+"/user/:id/application/", middleware.Chain(c.Applications, config.Read|config.ReadUser))
	app.Post(prefix+"/user/:id/application/", middleware.Chain(c.LinkApplications, config.Write|config.WriteUser))
	app.Delete(prefix+"/user/:id/application/", middleware.Chain(c.UnLinkApplications, config.Write|config.WriteUser))
	app.Get(prefix+"/user/:id/application/:applicationID", middleware.Chain(c.Application, config.Read|config.ReadUser))
	app.Put(prefix+"/user/:id/application/:applicationID", middleware.Chain(c.UpdateApplication, config.Write|config.WriteUser))
	app.Get(prefix+"/user/:id/role/", middleware.Chain(c.Roles, config.Read|config.ReadUser))
	app.Post(prefix+"/user/:id/role/", middleware.Chain(c.LinkRoles, config.Write|config.WriteUser))
	app.Delete(prefix+"/user/:id/role/", middleware.Chain(c.UnLinkRoles, config.Write|config.WriteUser))
}

```

### Full Example: Controller
```go
package controller

import (
	"sync"

	"app.gostartkit.com/go/auth/model"
	"app.gostartkit.com/go/auth/proxy"
	"app.gostartkit.com/go/auth/validator"
	"pkg.gostartkit.com/web"
)

var (
	_userController     *UserController
	_onceUserController sync.Once
)

// CreateUserController return *UserController
func CreateUserController() *UserController {

	_onceUserController.Do(func() {
		_userController = &UserController{}
	})

	return _userController
}

// UserController struct
type UserController struct {
}

// Index get users
func (r *UserController) Index(c *web.Ctx) (any, error) {

	filter := c.QueryFilter()
	orderBy := c.QueryOrderBy()
	page := c.QueryPage(_defaultPage)
	limit := c.QueryLimit(_defaultPageSize)

	return proxy.GetUsers(filter, orderBy, page, limit)
}

// Detail get user
func (r *UserController) Detail(c *web.Ctx) (any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := validator.Uint64("id", id); err != nil {
		return nil, err
	}

	return proxy.GetUser(id)
}

// CreateID create user.ID
func (r *UserController) CreateID(c *web.Ctx) (any, error) {
	return proxy.CreateUserId()
}

// Create create user
func (r *UserController) Create(c *web.Ctx) (any, error) {

	user := model.CreateUser()

	if err := c.TryParseBody(user); err != nil {
		return nil, err
	}

	if err := validator.CreateUser(user); err != nil {
		return nil, err
	}

	if _, err := proxy.CreateUser(user); err != nil {
		return nil, err
	}

	return user.ID, nil
}

// Update update user
func (r *UserController) Update(c *web.Ctx) (any, error) {

	var err error

	user := model.CreateUser()

	if err = c.TryParseBody(user); err != nil {
		return nil, err
	}

	if user.ID, err = c.ParamUint64("id"); err != nil {
		return nil, err
	}

	if err = validator.UpdateUser(user); err != nil {
		return nil, err
	}

	return proxy.UpdateUser(user)
}

// Patch update user
func (r *UserController) Patch(c *web.Ctx) (any, error) {

	attrs := c.HeaderAttrs()

	if err := validator.Int(web.HeaderAttrs, len(attrs)); err != nil {
		return nil, err
	}

	var err error

	user := model.CreateUser()

	if err = c.TryParseBody(user); err != nil {
		return nil, err
	}

	if user.ID, err = c.ParamUint64("id"); err != nil {
		return nil, err
	}

	if err = validator.PatchUser(user, attrs...); err != nil {
		return nil, err
	}

	return proxy.PatchUser(user, attrs...)
}

// UpdateStatus update user.Status
func (r *UserController) UpdateStatus(c *web.Ctx) (any, error) {

	var err error

	user := model.CreateUser()

	if err = c.TryParseBody(user); err != nil {
		return nil, err
	}

	if user.ID, err = c.ParamUint64("id"); err != nil {
		return nil, err
	}

	if err = validator.UpdateUserStatus(user); err != nil {
		return nil, err
	}

	return proxy.UpdateUserStatus(user)
}

// Destroy delete user
func (r *UserController) Destroy(c *web.Ctx) (any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := validator.Uint64("id", id); err != nil {
		return nil, err
	}

	return proxy.DestroyUserSoft(id)
}

// Applications return *model.ApplicationCollection, error
func (r *UserController) Applications(c *web.Ctx) (any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	filter := c.QueryFilter()
	orderBy := c.QueryOrderBy()
	page := c.QueryPage(_defaultPage)
	limit := c.QueryLimit(_defaultPageSize)

	return proxy.GetApplicationsByUserId(id, filter, orderBy, page, limit)
}

// LinkApplications return rowsAffected int64, error
func (r *UserController) LinkApplications(c *web.Ctx) (any, error) {

	var (
		applicationID []uint64
	)

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := c.TryParseBody(&applicationID); err != nil {
		return nil, err
	}

	return proxy.LinkUserApplications(id, applicationID...)
}

// UnLinkApplications return rowsAffected int64, error
func (r *UserController) UnLinkApplications(c *web.Ctx) (any, error) {

	var (
		applicationID []uint64
	)

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := c.TryParseBody(&applicationID); err != nil {
		return nil, err
	}

	return proxy.UnLinkUserApplications(id, applicationID...)
}

// Application return *model.ApplicationUser, error
func (r *UserController) Application(c *web.Ctx) (any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	applicationID, err := c.ParamUint64("applicationID")

	if err != nil {
		return nil, err
	}

	return proxy.GetUserApplication(id, applicationID)
}

// UpdateApplication return rowsAffected int64, error
func (r *UserController) UpdateApplication(c *web.Ctx) (any, error) {

	applicationUser := model.CreateApplicationUser()

	if err := c.TryParseBody(applicationUser); err != nil {
		return nil, err
	}

	var err error

	applicationUser.ApplicationID, err = c.ParamUint64("applicationID")

	if err != nil {
		return nil, err
	}

	applicationUser.UserId, err = c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := validator.UpdateUserApplication(applicationUser); err != nil {
		return nil, err
	}

	return proxy.UpdateUserApplication(applicationUser)
}

// Roles return *model.RoleCollection, error
func (r *UserController) Roles(c *web.Ctx) (any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	filter := c.QueryFilter()
	orderBy := c.QueryOrderBy()
	page := c.QueryPage(_defaultPage)
	limit := c.QueryLimit(_defaultPageSize)

	return proxy.GetRolesByUserId(id, filter, orderBy, page, limit)
}

// LinkRoles return rowsAffected int64, error
func (r *UserController) LinkRoles(c *web.Ctx) (any, error) {

	var (
		roleID []uint64
	)

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := c.TryParseBody(&roleID); err != nil {
		return nil, err
	}

	return proxy.LinkUserRoles(id, roleID...)
}

// UnLinkRoles return rowsAffected int64, error
func (r *UserController) UnLinkRoles(c *web.Ctx) (any, error) {

	var (
		roleID []uint64
	)

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := c.TryParseBody(&roleID); err != nil {
		return nil, err
	}

	return proxy.UnLinkUserRoles(id, roleID...)
}

```

### Notes

- Best performance for param/catch-all routing is achieved when params are pooled (already used in `Application`).
- For binary/avro responses, prefer returning `[]byte` or implementing `web.AvroMarshaler` to avoid extra encoding overhead.
- `TryParseBody` currently supports JSON/GOB/XML only.

### Acknowledgments

Thanks to all open-source projects, I’ve learned a lot from them.

Special thanks to：

- [httprouter](https://github.com/julienschmidt/httprouter): A high-performance HTTP router that inspired the routing logic in this project.
- [web](https://github.com/hoisie/web): A lightweight web framework that provided insights into efficient server design.
