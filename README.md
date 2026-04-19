# Web.go The library for web

中文文档: [README_CN.md](./README_CN.md)

### Performance First

This library is optimized around low-latency request handling, tight routing, and low-allocation parsing/writing paths.

Current benchmark snapshot on `darwin/arm64` (`Apple M2`):

<!-- BENCHMARK_SNAPSHOT:BEGIN -->
| Benchmark | Result | Memory |
|---|---:|---:|
| `BenchmarkServeHTTPStaticJSON` | `152.4 ns/op` | `16 B/op`, `1 alloc/op` |
| `BenchmarkServeHTTPPathParamJSON` | `196.3 ns/op` | `24 B/op`, `2 alloc/op` |
| `BenchmarkServeHTTPStaticJSONRawMessage` | `119.9 ns/op` | `40 B/op`, `2 alloc/op` |
| `BenchmarkTryParseJSONBodyFast` | `1417.0 ns/op` | `5600 B/op`, `20 alloc/op` |
| `BenchmarkPostBytes` | `38264.0 ns/op` | `6165 B/op`, `74 alloc/op` |
| `BenchmarkDoReqWithClientRawBody` | `189.4 ns/op` | `328 B/op`, `7 alloc/op` |
| `BenchmarkServeHTTPBinary` | `125.2 ns/op` | `40 B/op`, `2 alloc/op` |
| `BenchmarkServeHTTPAvro` | `124.7 ns/op` | `40 B/op`, `2 alloc/op` |
| `BenchmarkTreeGetValueParamPooled` | `14.2 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkTryParseIntSlice` | `121.2 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkTryParseStringSlice` | `34.9 ns/op` | `0 B/op`, `0 alloc/op` |
<!-- BENCHMARK_SNAPSHOT:END -->

Notes:

- Static JSON responses are down to a single allocation on the request path.
- Param and catch-all routing become `0 alloc` when params are pooled, which is already how `Application` runs.
- Pre-encoded JSON (`json.RawMessage`) has a dedicated write fast path.
- `TryParseJSONBodyFast` is the opt-in fast path for JSON request bodies when unknown-field rejection is not required.
- Client response decoding has an explicit raw-body fast path via `*web.RawBody`.
- Binary and avro responses have direct fast paths.
- Slice parsing hot paths avoid `strings.Split` and now run with `0 alloc`.

### Benchmark Workflow

Run the current benchmark suite:

```bash
go test -run '^$' -bench 'Benchmark(ServeHTTP|TreeGetValue|TryParse|TryInt|TryUint|TryBool|Post(JSON|Bytes)|DoReqWithClient(Struct|RawBody)|CtxWriteBinaryReader)' -benchmem ./...
```

Compare current results against the committed baseline:

```bash
./bench/compare.sh
```

Refresh the committed baseline:

```bash
./bench/update_baseline.sh
```

Generate a Markdown benchmark snapshot ready to paste into the README:

```bash
./bench/snapshot.sh
```

Update the benchmark snapshot blocks in `README.md` and `README_CN.md`:

```bash
./bench/update_snapshot_readme.sh
```

Useful overrides:

```bash
COUNT=3 ./bench/compare.sh
BENCH_EXPR='BenchmarkServeHTTP(StaticJSON|PathParamJSON)$' ./bench/compare.sh
CURRENT_FILE=./bench/servehttp.txt COUNT=3 ./bench/compare.sh
SHOW_MISSING=1 ./bench/compare.sh
COUNT=3 ./bench/update_baseline.sh
COUNT=3 ./bench/snapshot.sh
COUNT=3 ./bench/update_snapshot_readme.sh
```

Files:

- baseline: [bench/baseline.txt](./bench/baseline.txt)
- compare script: [bench/compare.sh](./bench/compare.sh)
- update script: [bench/update_baseline.sh](./bench/update_baseline.sh)
- snapshot script: [bench/snapshot.sh](./bench/snapshot.sh)
- README update script: [bench/update_snapshot_readme.sh](./bench/update_snapshot_readme.sh)

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
  - `Get`, `Post`, `Put`, `Patch`, `Delete`, `Head`, `Options`, `Handle`
- framework composition:
  - `Use`, `Group`, `SetErrorHandler`, `RegisterReader`, `RegisterWriter`
- server lifecycle:
  - `ListenAndServe`, `ListenAndServeTLS`, `Shutdown`
- helpers:
  - `ServeFiles`, `Redirect`, `TryParse(...)`, `TryXxx(...)`, `JSONErrorHandler`
- context (`*Ctx`) common methods:
  - request: `Method`, `Path`, `Query`, `Param`, `Body`, `ContentType`, `BearerToken`, `RequestID`
  - parse: `TryParseBody`, `TryParseJSONBodyFast`, `TryParseParam`, `TryParseQuery`, `TryParseForm`
  - response: `SetHeader`, `SetCookie`, `AllowCredentials`, content negotiation via `Accept`

### API Quick Reference (EN)

| Area | API | Description |
|---|---|---|
| Application | `New()` | Create app instance |
| Application | `Get/Post/Put/Patch/Delete/Head/Options(path, handler)` | Register route handler |
| Application | `Handle(method, path, handler)` | Register route handler for an arbitrary HTTP method |
| Application | `Use(middleware...)` | Apply app-level middleware to subsequently registered routes |
| Application | `Group(prefix, middleware...)` | Create route groups with shared prefix and middleware |
| Application | `SetErrorHandler(handler)` | Install a custom route error handler |
| Application | `RegisterReader(contentType, reader)` | Override request decoding for a media type |
| Application | `RegisterWriter(contentType, writer)` | Override response encoding for a media type |
| Application | `ServeFiles("/static/*filepath", fs)` | Serve static files with catch-all path |
| Application | `ListenAndServe(network, addr, ...opts)` | Start HTTP server |
| Application | `ListenAndServeTLS(network, addr, tlsConfig, ...opts)` | Start HTTPS server |
| Application | `Shutdown(ctx)` | Graceful shutdown |
| Context | `Param(name)`, `Query(name)`, `Form(name)`, `RequestID()` | Read path/query/form values and middleware-provided request ID |
| Context | `TryParseBody(v)` | Parse request body by content type (JSON/GOB/XML) |
| Context | `TryParseJSONBodyFast(v)` | Fast JSON body parse using pooled buffer + `json.Unmarshal` |
| Context | `TryParseParam/Query/Form(name, &v)` | Parse string values into typed value |
| Context | `SetHeader`, `SetCookie`, `SetContentType`, `SetStatus` | Write response headers and override the default success status |
| Context | `Request()`, `ResponseWriter()`, `Context()` | Access raw HTTP objects |
| Middleware | `RequestID`, `Recover`, `RecoverWithOptions`, `Timeout`, `AccessLog`, `AccessLogWithOptions` | Built-in opt-in middleware helpers |
| Client | `Get/Post/Put/Patch/Delete/Do` | HTTP client helpers using `http.DefaultClient` |
| Client | `GetWithClient/PostWithClient/PutWithClient/PatchWithClient/DeleteWithClient/DoWithClient` | HTTP helpers with explicit `*http.Client` |
| Client | `DoReq/DoReqWithClient` | Execute prepared requests and decode JSON or `RawBody` responses |
| Client | `PostBytes/PutBytes/PatchBytes/DoBytes` | Send pre-encoded request bodies without JSON encoding |
| Client | `PostBytesWithClient/PutBytesWithClient/PatchBytesWithClient/DoBytesWithClient` | Pre-encoded body helpers with explicit `*http.Client` |
| Client | `TryGet/TryPost/TryPut/TryPatch/TryDelete/TryDo` | HTTP helpers with retry loop |
| Client | `TryGetWithClient/TryPostWithClient/TryPutWithClient/TryPatchWithClient/TryDeleteWithClient/TryDoWithClient` | Retry helpers with explicit `*http.Client` |
| Client | `TryPostBytes/TryPutBytes/TryPatchBytes/TryDoBytes` | Retry-capable helpers for pre-encoded request bodies |
| Client | `TryPostBytesWithClient/TryPutBytesWithClient/TryPatchBytesWithClient/TryDoBytesWithClient` | Retry-capable pre-encoded helpers with explicit `*http.Client` |
| Error | `NewErr(code, msg)` | Error with HTTP status code |
| Error | `Redirect(url, code)` | Return redirect response from handler |
| Error | `JSONErrorHandler(includeRequestID)` | Write structured JSON API errors |

### Response Behavior

- Handler return value controls response:
  - `(nil, nil)` -> `204 No Content`
  - `(value, nil)` -> `200 OK`
  - call `c.SetStatus(code)` to explicitly override the default success status
  - `(_, err)` -> status code from framework error type, body contains `err.Error()`
- Response format is selected by request `Accept` header:
  - `application/json`
  - `application/x-gob`
  - `application/xml`
  - `application/octet-stream`
  - `application/x-avro`

### Modern Framework Features

- Middleware and route groups are registration-time features:
  - `app.Use(...)`
  - `app.Group("/api", ...)`
  - group-local `Use(...)`
- Built-in middleware is explicit opt-in:
  - `RequestID`
  - `Recover`
  - `RecoverWithOptions`
  - `Timeout`
  - `AccessLog`
  - `AccessLogWithOptions`
- Structured API errors are opt-in via `SetErrorHandler(JSONErrorHandler(...))`
- Reader/writer overrides are media-type specific and do not affect the default hot path unless registered

```go
app := web.New()
app.Use(web.RequestID("", nil), web.Recover(nil))
app.SetErrorHandler(web.JSONErrorHandler(true))

api := app.Group("/api", web.Timeout(2*time.Second))
api.Get("/users/:id", func(c *web.Ctx) (any, error) {
	return map[string]string{
		"id":         c.Param("id"),
		"request_id": c.RequestID(),
	}, nil
})
```

For finer control, use the options-based middleware variants:

```go
app.Use(
	web.RecoverWithOptions(web.RecoverOptions{
		DefaultStatus: http.StatusServiceUnavailable,
		DefaultBody:   "UNAVAILABLE",
	}),
	web.AccessLogWithOptions(web.AccessLogOptions{
		Log: func(c *web.Ctx, entry web.AccessLogEntry) {
			// route-aware access logging hook
		},
	}),
)
```

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
  - `DoReq` / `DoReqWithClient` now recognize `*web.RawBody`
  - existing JSON destinations like `[]byte` and `json.RawMessage` keep their original JSON semantics
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

Use `DoReqWithClient` with `*web.RawBody` when you want the response payload without JSON decoding cost.

```go
req, _ := http.NewRequest(http.MethodGet, "https://example.com/data", nil)

var raw web.RawBody
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
