// Package web implements a performance-oriented HTTP framework for Go.
//
// The package is designed around a small, explicit core:
//
//   - tree-based routing for static, parameter, and catch-all paths
//   - pooled request context and pooled route params
//   - explicit request parsing and response writing
//   - opt-in middleware, route groups, and structured error handling
//   - integrated HTTP client helpers with retry and raw-body fast paths
//
// The default request path is intentionally tight. Additional framework features
// such as middleware, custom readers/writers, structured JSON error output, and
// client transport tuning are all opt-in so that unused capabilities do not add
// work to the hot path.
//
// Quick start:
//
//	package main
//
//	import (
//		"log"
//
//		"pkg.gostartkit.com/web"
//	)
//
//	func main() {
//		app := web.New()
//
//		app.Get("/health", func(c *web.Ctx) (any, error) {
//			return map[string]string{"status": "ok"}, nil
//		})
//
//		log.Fatal(app.ListenAndServe("tcp", ":8080"))
//	}
//
// Modern framework composition:
//
//	app := web.New()
//	app.Use(web.RequestID("", nil), web.Recover(nil))
//	app.SetErrorHandler(web.JSONErrorHandler(true))
//
//	api := app.Group("/api", web.Timeout(2*time.Second))
//	api.Get("/users/:id", func(c *web.Ctx) (any, error) {
//		return map[string]string{
//			"id":         c.Param("id"),
//			"request_id": c.RequestID(),
//		}, nil
//	})
//
// Request and response model:
//
// Handlers use the form:
//
//	func(c *web.Ctx) (any, error)
//
// The returned value controls the default response semantics:
//
//   - (nil, nil) -> 204 No Content
//   - (value, nil) -> 200 OK, or 201 Created for POST
//   - (_, err) -> status inferred from framework error type
//
// Request bodies are parsed from Content-Type using Ctx.TryParseBody, or
// Ctx.TryParseJSONBodyFast when unknown-field rejection is not required.
//
// Responses are negotiated from Accept and support JSON, GOB, XML, binary, and Avro.
// Pre-encoded JSON can be returned as json.RawMessage. Raw client response bytes use
// the explicit RawBody type.
//
// Performance guidance:
//
//   - Prefer []byte or AvroMarshaler for binary and Avro output
//   - Prefer PostBytes/PutBytes/PatchBytes/DoBytes when request payloads are already encoded
//   - Reuse destination slices in TryParse hot paths
//   - Use explicit *WithClient helpers when transport-level tuning matters
//
// See README.md and README_CN.md for benchmark snapshots, compatibility notes,
// and the full API surface.
package web
