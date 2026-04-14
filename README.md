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

### Notes

- Best performance for param/catch-all routing is achieved when params are pooled (already used in `Application`).
- For binary/avro responses, prefer returning `[]byte` or implementing `web.AvroMarshaler` to avoid extra encoding overhead.
- `TryParseBody` currently supports JSON/GOB/XML only.

### Acknowledgments

Thanks to all open-source projects, I’ve learned a lot from them.

Special thanks to：

- [httprouter](https://github.com/julienschmidt/httprouter): A high-performance HTTP router that inspired the routing logic in this project.
- [web](https://github.com/hoisie/web): A lightweight web framework that provided insights into efficient server design.
