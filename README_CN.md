# Web.go Web 开发库

English Version: [README.md](./README.md)

### 性能至上

本库围绕低延迟请求处理、紧凑的路由以及低分配的解析/写入路径进行了优化。

当前在 `darwin/arm64` (`Apple M2`) 上的基准测试快照：

| 基准测试 | 结果 | 内存 |
|---|---:|---:|
| `BenchmarkServeHTTPStaticJSON` | `195.1 ns/op` | `80 B/op`, `5 allocs/op` |
| `BenchmarkServeHTTPPathParamJSON` | `288.7 ns/op` | `128 B/op`, `6 allocs/op` |
| `BenchmarkServeHTTPBinary` | `171.7 ns/op` | `104 B/op`, `6 allocs/op` |
| `BenchmarkTreeGetValueStatic` | `2.616 ns/op` | `0 B/op`, `0 allocs/op` |
| `BenchmarkTreeGetValueParamPooled` | `14.06 ns/op` | `0 B/op`, `0 allocs/op` |
| `BenchmarkTryParseIntSlice` | `138.3 ns/op` | `80 B/op`, `1 alloc/op` |
| `BenchmarkTryParseStringSlice` | `59.62 ns/op` | `80 B/op`, `1 alloc/op` |

备注：

- 静态路由查找实际上是零分配的。
- 当参数被池化时，参数路由和通配路由变为 `0 alloc`，这已经是 `Application` 的运行方式。
- 二进制和 Avro 响应具有直接的快速路径。
- 切片解析热路径经过优化，避免了中间的 `strings.Split` 分配模式。

### 基准测试流程

运行当前的基准测试套件：

```bash
go test -run '^$' -bench 'Benchmark(ServeHTTP|TreeGetValue|TryParse|TryInt|TryUint|TryBool|PostJSON|CtxWriteBinaryReader)' -benchmem ./...
```

将当前结果与提交的基准线进行比较：

```bash
./bench/compare.sh
```

文件：

- 基准线: [bench/baseline.txt](./bench/baseline.txt)
- 比较脚本: [bench/compare.sh](./bench/compare.sh)

### 性能指南

- 对于二进制/Avro 响应，首选 `[]byte` 或 `web.AvroMarshaler`。
- 在热路径中调用 `TryParse(..., &slice)` 时，重用目标切片。
- 如果孤立地对路由进行基准测试，请首选池化参数路径；框架在正常请求处理中已经这样做了。
- 将单次基准测试运行视为存在噪声。使用基准线比较脚本作为方向，而不是凭直觉。

### 快速入门

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

### API 索引

- `web.New() *Application`
- 路由注册：
  - `Get`, `Post`, `Put`, `Patch`, `Delete`, `Head`, `Options`
- 服务器生命周期：
  - `ListenAndServe`, `ListenAndServeTLS`, `Shutdown`
- 辅助函数：
  - `ServeFiles`, `Redirect`, `TryParse(...)`, `TryXxx(...)`
- 上下文 (`*Ctx`) 常用方法：
  - 请求：`Method`, `Path`, `Query`, `Param`, `Body`, `ContentType`, `BearerToken`
  - 解析：`TryParseBody`, `TryParseParam`, `TryParseQuery`, `TryParseForm`
  - 响应：`SetHeader`, `SetCookie`, `AllowCredentials`, 通过 `Accept` 进行内容协商

### API 快速参考 (CN)

| 领域 | API | 描述 |
|---|---|---|
| 应用程序 | `New()` | 创建应用程序实例 |
| 应用程序 | `Get/Post/Put/Patch/Delete/Head/Options(path, handler)` | 注册路由处理器 |
| 应用程序 | `ServeFiles("/static/*filepath", fs)` | 使用通配路径提供静态文件服务 |
| 应用程序 | `ListenAndServe(network, addr, ...opts)` | 启动 HTTP 服务器 |
| 应用程序 | `ListenAndServeTLS(network, addr, tlsConfig, ...opts)` | 启动 HTTPS 服务器 |
| 应用程序 | `Shutdown(ctx)` | 优雅关闭 |
| 上下文 | `Param(name)`, `Query(name)`, `Form(name)` | 读取路径/查询/表单值 |
| 上下文 | `TryParseBody(v)` | 根据内容类型（JSON/GOB/XML）解析请求体 |
| 上下文 | `TryParseParam/Query/Form(name, &v)` | 将字符串值解析为类型化值 |
| 上下文 | `SetHeader`, `SetCookie`, `SetContentType` | 写入响应头 |
| 上下文 | `Request()`, `ResponseWriter()`, `Context()` | 访问原始 HTTP 对象 |
| 客户端 | `Get/Post/Put/Patch/Delete/Do` | HTTP 客户端辅助函数 |
| 客户端 | `TryGet/TryPost/TryPut/TryPatch/TryDelete/TryDo` | 带有重试循环的 HTTP 辅助函数 |
| 错误 | `NewErr(code, msg)` | 带有 HTTP 状态码的错误 |
| 错误 | `Redirect(url, code)` | 从处理器返回重定向响应 |

### 响应行为

- 处理器返回值控制响应：
  - `(nil, nil)` -> `204 No Content`
  - `(value, nil)` -> `200 OK` (`POST` 使用 `201 Created`)
  - `(_, err)` -> 状态码来自框架错误类型，响应体包含 `err.Error()`
- 响应格式通过请求的 `Accept` 头部选择：
  - `application/json`
  - `application/x-gob`
  - `application/xml`
  - `application/octet-stream`
  - `application/x-avro`

### 兼容性 / 破坏性变更

- `Try*` 重试语义更新：
  - `retry <= 0` 现在仍执行一次请求尝试。
  - 对于 `ErrUnauthorized`、`ErrForbidden` 和 `ErrBadRequest`（包括包装后的），重试循环会提早停止。
- `TryDo` 现在支持跨重试的请求体安全重放（请求体会被缓冲一次并在每次尝试时重新创建）。
- 实现了 `Ctx.writeBinary` 和 `Ctx.writeAvro`：
  - 之前这些媒体类型的行为是 `ErrNotImplemented`。
  - 现在它们支持快速路径直接写入（见二进制 / Avro 响应章节）。
- 重定向用法：
  - 仅返回 `ErrMovedPermanently` 不会设置 `Location`。
  - 使用 `web.Redirect(url, code)` 生成正确的重定向响应头。
- 头部协商改进：
  - 带有参数的 `Accept`/`Content-Type` 值（例如 `application/json; charset=utf-8`）现在可以被正确解析。

迁移建议：

- 如果你依赖 `retry=0` 来跳过外部调用，请在调用方替换为显式的条件判断。
- 如果你的处理器使用了 `application/octet-stream` 或 `application/x-avro`，你现在可以直接返回 `[]byte`、`io.Reader` 或自定义的序列化类型。
- 对于重定向，请迁移到 `web.Redirect(...)` 以获得可预测的行为。

### 当前功能 (2026-04)

- 路由：
  - 静态路径, `:param`, `*catchAll`
  - 高性能树匹配器（灵感来自 `httprouter`）
- 根据 `Accept` 进行响应编码：
  - `application/json`
  - `application/x-gob`
  - `application/xml`
  - `application/octet-stream` (已实现)
  - `application/x-avro` (已实现)
- 根据 `Content-Type` 进行请求体解析：
  - `application/json`
  - `application/x-gob`
  - `application/xml`

### 二进制 / Avro 响应

`Ctx.writeBinary` 和 `Ctx.writeAvro` 针对快速路径进行了优化。

- 二进制快速路径输入类型：
  - `[]byte`
  - `string`
  - `*bytes.Buffer`
  - `io.Reader`
  - `encoding.BinaryMarshaler`
- Avro 快速路径输入类型：
  - `web.AvroMarshaler`
  - 对于上述相同的输入类型，会回退到二进制写入器

```go
type Event struct {
	Raw []byte
}

func (e Event) MarshalAvro() ([]byte, error) {
	return e.Raw, nil
}

app.Get("/payload", func(c *web.Ctx) (any, error) {
	// 客户端发送: Accept: application/x-avro
	return Event{Raw: []byte{0xAA, 0xBB}}, nil
})
```

### 重定向辅助函数

使用 `web.Redirect(url, code)` 返回重定向响应。

```go
app.Get("/old", func(c *web.Ctx) (any, error) {
	return web.Redirect("/new", http.StatusMovedPermanently)
})
```

### HTTP 客户端重试行为

`TryGet`, `TryPost`, `TryPut`, `TryPatch`, `TryDelete`, `TryDo`:

- `retry <= 0` 仍执行至少 **一次** 请求。
- 对于非可重试错误会提早停止：
  - `ErrUnauthorized`
  - `ErrForbidden`
  - `ErrBadRequest` (包括包装后的)
- `TryDo` 安全地通过请求体重放进行重试（请求体被缓存一次并在每次尝试时重新创建）。

### 注意事项

- 参数/通配路由的最佳性能是在参数被池化时实现的（`Application` 中已使用）。
- 对于二进制/Avro 响应，首选返回 `[]byte` 或实现 `web.AvroMarshaler` 以避免额外的编码开销。
- `TryParseBody` 目前仅支持 JSON/GOB/XML。

### 致谢

感谢所有开源项目，我从中受益匪浅。

特别感谢：

- [httprouter](https://github.com/julienschmidt/httprouter): 一个高性能的 HTTP 路由，启发了本项目中的路由逻辑。
- [web](https://github.com/hoisie/web): 一个轻量级的 Web 框架，提供了关于高效服务器设计的见解。
