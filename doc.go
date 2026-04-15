// Package web implements a high-performance HTTP web framework for Go.
//
// The framework is optimized for low-latency request handling, efficient routing,
// and minimal allocations across the request lifecycle.
//
// Design Philosophy:
//
// The framework focuses on simplicity, performance, and explicit control.
// It avoids heavy abstractions while providing enough flexibility for building
// scalable HTTP services.
//
// Key Features:
//
//   - Provides high-performance routing with support for static, parameter (:name),
//     and catch-all (*path) routes using a tree-based matcher.
//   - Uses a specialized Ctx object for efficient request/response handling,
//     minimizing allocations via parameter pooling.
//   - Supports content negotiation for multiple formats (JSON, GOB, XML, Binary, Avro).
//   - Includes an integrated HTTP client with retry support and safe request replay.
//
// Quick Start:
//
//	package main
//
//	import (
//		"log"
//
//		"github.com/gostartkit/web"
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
// Routing and Parameters:
//
//	app.Get("/user/:id", func(c *web.Ctx) (any, error) {
//		id := c.Param("id")
//		return "User ID: " + id, nil
//	})
//
// Handler Signatures:
//
// In addition to the traditional handler:
//
//	app.Get("/users", func(c *web.Ctx) (any, error) { ... })
//
// The framework also supports automatic parameter injection:
//
//	type UserQuery struct {
//		Page  uint32 `query:"page"`
//		Limit uint32 `query:"limit"`
//	}
//
//	app.Get("/users", func(ctx context.Context, q UserQuery) (any, error) {
//		return q, nil
//	})
//
// Supported injected types include:
// - *web.Ctx
// - context.Context
// - struct (bound from query, body, path, or headers)
//
// Content Negotiation:
//
// Responses are encoded based on the "Accept" header.
// Supported formats include JSON (default), GOB, XML,
// binary streams, and Avro.
//
// Request bodies are decoded according to the "Content-Type" header
// using Ctx.TryParseBody.
//
// Performance Guidelines:
//
// - Prefer returning []byte or AvroMarshaler for zero-copy fast paths.
// - Reuse buffers and slices when parsing in hot paths.
// - Avoid unnecessary allocations inside handlers.
// - Parameter storage is pooled internally by the Application.
//
// For more details, see the README.md or visit the project repository.
package web
