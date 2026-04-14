# Web.go 中文文档

轻量高性能 Go Web 库，包含高效路由匹配、请求上下文封装、内容协商、HTTP 客户端辅助与重试能力。

## 快速开始

```go
package main

import (
	"log"

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

## 主要能力（2026-04）

- 路由：
  - 静态路由
  - `:param` 参数路由
  - `*catchAll` 捕获路由
- 响应编码（由 `Accept` 协商）：
  - `application/json`
  - `application/x-gob`
  - `application/xml`
  - `application/octet-stream`
  - `application/x-avro`
- 请求体解析（由 `Content-Type` 识别）：
  - `application/json`
  - `application/x-gob`
  - `application/xml`
- HTTP 客户端辅助：
  - `Get/Post/Put/Patch/Delete/Do`
  - `TryGet/TryPost/TryPut/TryPatch/TryDelete/TryDo`（带重试）

## API 速查

### Application

- `web.New() *Application`
- `Get/Post/Put/Patch/Delete/Head/Options(path, handler)`
- `ServeFiles("/static/*filepath", fs)`
- `ListenAndServe(network, addr, ...opts)`
- `ListenAndServeTLS(network, addr, tlsConfig, ...opts)`
- `Shutdown(ctx)`

### Context (`*Ctx`)

- 请求信息：
  - `Method`, `Path`, `Host`, `RemoteAddr`, `Request`, `Context`
- 参数读取：
  - `Param`, `Query`, `Form`, `PostForm`
- 参数解析：
  - `TryParseBody`
  - `TryParseParam`, `TryParseQuery`, `TryParseForm`
  - `ParamInt/ParamUint/...`, `QueryInt/QueryUint/...`, `FormInt/FormUint/...`
- 响应相关：
  - `SetHeader`, `SetContentType`, `SetCookie`, `AllowCredentials`

### 错误与重定向

- `NewErr(code, msg)`：创建带 HTTP 状态码的错误
- `Redirect(url, code)`：在 handler 中返回标准重定向响应

## 响应行为说明

- Handler 返回 `(nil, nil)`：`204 No Content`
- Handler 返回 `(value, nil)`：默认 `200 OK`，`POST` 为 `201 Created`
- Handler 返回 `(_, err)`：状态码来自框架错误类型，响应体写入 `err.Error()`

## Binary / Avro 输出

`Ctx.writeBinary` 与 `Ctx.writeAvro` 已实现高性能快路径。

### Binary 支持类型

- `[]byte`
- `string`
- `*bytes.Buffer`
- `io.Reader`
- `encoding.BinaryMarshaler`

### Avro 支持类型

- `web.AvroMarshaler`
- 不满足时回退到 Binary 写出逻辑

```go
type Event struct {
	Raw []byte
}

func (e Event) MarshalAvro() ([]byte, error) {
	return e.Raw, nil
}

app.Get("/event", func(c *web.Ctx) (any, error) {
	// 客户端请求头: Accept: application/x-avro
	return Event{Raw: []byte{0xAA, 0xBB}}, nil
})
```

## HTTP 客户端重试语义

`TryGet`, `TryPost`, `TryPut`, `TryPatch`, `TryDelete`, `TryDo`：

- `retry <= 0` 也会至少执行一次请求
- 以下错误会提前停止重试：
  - `ErrUnauthorized`
  - `ErrForbidden`
  - `ErrBadRequest`（包含被 wrap 的情况）
- `TryDo` 支持请求体重放（缓存一次 body，每次重试重建 reader）

## 兼容性 / 行为变更

- `application/octet-stream` 与 `application/x-avro` 之前为未实现，现在已支持输出
- 仅返回 `ErrMovedPermanently` 不会自动写 `Location`，请使用 `web.Redirect(url, code)`
- `Accept/Content-Type` 带参数（例如 `application/json; charset=utf-8`）可正确识别

迁移建议：

- 若历史逻辑依赖 `retry=0` 跳过请求，请在调用方显式判断
- 需要二进制/Avro 输出时，优先返回 `[]byte` 或实现 `AvroMarshaler`

## 性能基准

运行基准：

```bash
go test -run '^$' -bench 'Benchmark(ServeHTTP|TreeGetValue|TryParseBody|PostJSON)' -benchmem ./...
```

Tree 专项基准（路由匹配本体）：

```bash
go test -run '^$' -bench 'BenchmarkTreeGetValue(Static|Param|CatchAll|ParamPooled|CatchAllPooled)' -benchmem ./...
```

## 致谢

- [httprouter](https://github.com/julienschmidt/httprouter)
- [web](https://github.com/hoisie/web)
