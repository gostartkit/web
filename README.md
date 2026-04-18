# Web.go The library for web

中文文档: [README_CN.md](./README_CN.md)

### Performance First

This library is optimized around low-latency request handling, tight routing, and low-allocation parsing/writing paths.

Current benchmark snapshot on `darwin/arm64` (`Apple M2`):

| Benchmark | Result | Memory |
|---|---:|---:|
| `BenchmarkServeHTTPStaticJSON` | `157.9 ns/op` | `16 B/op`, `1 alloc/op` |
| `BenchmarkServeHTTPPathParamJSON` | `202.7 ns/op` | `24 B/op`, `2 allocs/op` |
| `BenchmarkServeHTTPStaticJSONRawMessage` | `124.1 ns/op` | `40 B/op`, `2 allocs/op` |
| `BenchmarkTryParseJSONBodyFast` | `1413 ns/op` | `5599 B/op`, `20 allocs/op` |
| `BenchmarkPostBytes` | `38179 ns/op` | `6169 B/op`, `74 allocs/op` |
| `BenchmarkDoReqWithClientBytes` | `192.8 ns/op` | `328 B/op`, `7 allocs/op` |
| `BenchmarkServeHTTPBinary` | `197.1 ns/op` | `40 B/op`, `2 allocs/op` |
| `BenchmarkServeHTTPAvro` | `144.7 ns/op` | `40 B/op`, `2 allocs/op` |
| `BenchmarkTreeGetValueParamPooled` | `14.29 ns/op` | `0 B/op`, `0 allocs/op` |
| `BenchmarkTryParseIntSlice` | `98.10 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkTryParseStringSlice` | `36.58 ns/op` | `0 B/op`, `0 alloc/op` |

Notes:

- Static JSON responses are down to a single allocation on the request path.
- Param and catch-all routing become `0 alloc` when params are pooled, which is already how `Application` runs.
- Pre-encoded JSON (`json.RawMessage`) has a dedicated write fast path.
- `TryParseJSONBodyFast` is the opt-in fast path for JSON request bodies when unknown-field rejection is not required.
- Client response decoding has a raw-body fast path for `*[]byte`, `*json.RawMessage`, and `*bytes.Buffer`.
- Binary and avro responses have direct fast paths.
- Slice parsing hot paths avoid `strings.Split` and now run with `0 alloc`.

### Benchmark Workflow

Run the current benchmark suite:

```bash
go test -run '^$' -bench 'Benchmark(ServeHTTP|TreeGetValue|TryParse|TryInt|TryUint|TryBool|Post(JSON|Bytes)|DoReqWithClient(Struct|Bytes)|CtxWriteBinaryReader)' -benchmem ./...
```

Compare current results against the committed baseline:

```bash
./bench/compare.sh
```

Files:

- baseline: [bench/baseline.txt](./bench/baseline.txt)
- compare script: [bench/compare.sh](./bench/compare.sh)

### Performance Guidelines

- Prefer `[]byte` or `web.AvroMarshaler` for binary/avro responses.
- Prefer `PostBytes/PutBytes/PatchBytes/DoBytes` when the request body is already encoded.
- Prefer `*WithClient` helpers when you need tuned timeouts, connection pooling, or a custom transport.
- Reuse destination slices when calling `TryParse(..., &slice)` in hot paths.
- Prefer pooled param paths if you benchmark routing in isolation; the framework already does this in normal request handling.
- Treat single benchmark runs as noisy. Use the baseline comparison script for direction, not intuition.

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
  - parse: `TryParseBody`, `TryParseJSONBodyFast`, `TryParseParam`, `TryParseQuery`, `TryParseForm`
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
| Context | `TryParseJSONBodyFast(v)` | Fast JSON body parse using pooled buffer + `json.Unmarshal` |
| Context | `TryParseParam/Query/Form(name, &v)` | Parse string values into typed value |
| Context | `SetHeader`, `SetCookie`, `SetContentType` | Write response headers |
| Context | `Request()`, `ResponseWriter()`, `Context()` | Access raw HTTP objects |
| Client | `Get/Post/Put/Patch/Delete/Do` | HTTP client helpers using `http.DefaultClient` |
| Client | `GetWithClient/PostWithClient/PutWithClient/PatchWithClient/DeleteWithClient/DoWithClient` | HTTP helpers with explicit `*http.Client` |
| Client | `DoReq/DoReqWithClient` | Execute prepared requests and decode JSON/raw response bodies |
| Client | `PostBytes/PutBytes/PatchBytes/DoBytes` | Send pre-encoded request bodies without JSON encoding |
| Client | `PostBytesWithClient/PutBytesWithClient/PatchBytesWithClient/DoBytesWithClient` | Pre-encoded body helpers with explicit `*http.Client` |
| Client | `TryGet/TryPost/TryPut/TryPatch/TryDelete/TryDo` | HTTP helpers with retry loop |
| Client | `TryGetWithClient/TryPostWithClient/TryPutWithClient/TryPatchWithClient/TryDeleteWithClient/TryDoWithClient` | Retry helpers with explicit `*http.Client` |
| Client | `TryPostBytes/TryPutBytes/TryPatchBytes/TryDoBytes` | Retry-capable helpers for pre-encoded request bodies |
| Client | `TryPostBytesWithClient/TryPutBytesWithClient/TryPatchBytesWithClient/TryDoBytesWithClient` | Retry-capable pre-encoded helpers with explicit `*http.Client` |
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
- Raw body helpers added:
  - `PostBytes`, `PutBytes`, `PatchBytes`, `DoBytes`
  - `TryPostBytes`, `TryPutBytes`, `TryPatchBytes`, `TryDoBytes`
  - default request headers are `Content-Type: application/octet-stream` and `Accept: application/json`
- Explicit client helpers added:
  - `DoReqWithClient`, `DoWithClient`, `DoBytesWithClient`
  - wrapper and retry variants for `Get/Post/Put/Patch/Delete`
  - use these when transport-level performance tuning matters
- Raw response fast path added:
  - `DoReq` / `DoReqWithClient` now recognize `*[]byte`, `*json.RawMessage`, and `*bytes.Buffer`
  - use these when the caller wants the response payload without JSON decoding cost
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

### Fast JSON Body Parse

Use `TryParseJSONBodyFast` when the request body is JSON and unknown-field rejection is not required.

```go
app.Post("/ingest", func(c *web.Ctx) (any, error) {
	var req struct {
		ID int `json:"id"`
	}

	if err := c.TryParseJSONBodyFast(&req); err != nil {
		return nil, err
	}

	return struct {
		Ok bool `json:"ok"`
	}{Ok: true}, nil
})
```

### Client Raw Response

Use `DoReqWithClient` with `*[]byte`, `*json.RawMessage`, or `*bytes.Buffer` when you want the response payload without JSON decoding cost.

```go
req, _ := http.NewRequest(http.MethodGet, "https://example.com/data", nil)

var raw []byte
if err := web.DoReqWithClient(client, req, &raw, nil); err != nil {
	panic(err)
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
