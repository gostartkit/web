# Web.go Web 开发库

English Version: [README.md](./README.md)

### 性能至上

本库围绕低延迟请求处理、紧凑的路由以及低分配的解析/写入路径进行了优化。

当前在 `darwin/arm64` (`Apple M2`) 上的基准测试快照：

<!-- BENCHMARK_SNAPSHOT:BEGIN -->
| Benchmark | Result | Memory |
|---|---:|---:|
| `BenchmarkServeHTTPStaticJSON` | `138.7 ns/op` | `16 B/op`, `1 alloc/op` |
| `BenchmarkServeHTTPPathParamJSON` | `182.9 ns/op` | `24 B/op`, `2 alloc/op` |
| `BenchmarkServeHTTPNoContent` | `19.8 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkServeHTTPManualWrite` | `21.4 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkServeHTTPStandardHandler` | `22.9 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkServeHTTPBlob` | `69.9 ns/op` | `16 B/op`, `1 alloc/op` |
| `BenchmarkServeHTTPStaticJSONRawMessage` | `109.2 ns/op` | `40 B/op`, `2 alloc/op` |
| `BenchmarkTryParseJSONBodyFast` | `1392.6 ns/op` | `5599 B/op`, `20 alloc/op` |
| `BenchmarkServeHTTPBinary` | `113.0 ns/op` | `40 B/op`, `2 alloc/op` |
| `BenchmarkServeHTTPAvro` | `113.1 ns/op` | `40 B/op`, `2 alloc/op` |
| `BenchmarkTreeGetValueParamPooled` | `14.7 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkCtxParamUint64` | `10.8 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkTryParseInt64` | `10.6 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkTryParseUint64` | `10.0 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkTryParseIntSlice` | `32.7 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkTryParseStringSlice` | `23.0 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkParseMediaTypeExactJSON` | `1.9 ns/op` | `0 B/op`, `0 alloc/op` |
| `BenchmarkAcceptMediaTypeEmpty` | `2.0 ns/op` | `0 B/op`, `0 alloc/op` |
<!-- BENCHMARK_SNAPSHOT:END -->

备注：

- `Memory` 来自 Go benchmark 的 `B/op` 和 `allocs/op`，该快照使用 `-benchmem` 采集。
- 静态 JSON 响应路径已经压到单次分配。
- 无内容响应和手写响应路径保持 `0 alloc`。
- 标准 `net/http` handler 可以以 `0 alloc` 的请求期开销挂载。
- `Ctx.Blob` 为预编码字节响应提供显式快速路径。
- 当参数被池化时，参数路由和通配路由变为 `0 alloc`，这已经是 `Application` 的运行方式。
- 预编码 JSON (`json.RawMessage`) 有独立的快速写出路径。
- `TryParseJSONBodyFast` 是 JSON 请求体的显式快路径，适用于不要求拒绝未知字段的场景。
- 客户端响应解码通过 `*web.RawBody` 提供显式的原始字节快速路径。
- 二进制和 Avro 响应具有直接的快速路径。
- 整数和切片解析热路径避免了额外扫描和中间切片，同时保持 `0 alloc`。

### 基准测试流程

运行当前的基准测试套件：

```bash
go test -run '^$' -bench 'Benchmark(ServeHTTP|TreeGetValue|TryParse|TryInt|TryUint|TryBool|Post(JSON|Bytes)|DoReqWithClient(Struct|RawBody)|Ctx|ParamsVal|ParseMediaType|AcceptMediaType)' -benchmem ./...
```

将当前结果与提交的基准线进行比较：

```bash
./bench/compare.sh
```

刷新提交到仓库的基准线：

```bash
./bench/update_baseline.sh
```

生成可直接粘贴到 README 的 Markdown 性能快照：

```bash
./bench/snapshot.sh
```

直接更新 `README.md` 和 `README_CN.md` 中的快照区块：

```bash
./bench/update_snapshot_readme.sh
```

常用覆盖方式：

```bash
COUNT=3 ./bench/compare.sh
BENCH_EXPR='BenchmarkServeHTTP(StaticJSON|PathParamJSON)$' ./bench/compare.sh
CURRENT_FILE=./bench/servehttp.txt COUNT=3 ./bench/compare.sh
SHOW_MISSING=1 ./bench/compare.sh
COUNT=3 ./bench/update_baseline.sh
COUNT=3 ./bench/snapshot.sh
COUNT=3 ./bench/update_snapshot_readme.sh
```

文件：

- 基准线: [bench/baseline.txt](./bench/baseline.txt)
- 比较脚本: [bench/compare.sh](./bench/compare.sh)
- 刷新脚本: [bench/update_baseline.sh](./bench/update_baseline.sh)
- 快照脚本: [bench/snapshot.sh](./bench/snapshot.sh)
- README 更新脚本: [bench/update_snapshot_readme.sh](./bench/update_snapshot_readme.sh)

### 性能指南

- 对于二进制/Avro 响应，首选 `[]byte` 或 `web.AvroMarshaler`。
- 对于需要立即写出的预编码字节响应，优先使用 `c.Blob(...)`。
- 集成已有 `net/http` handler 时使用 `HandleHTTP`/`GetHTTP`/`PostHTTP`。
- 如果请求体已经编码完成，优先使用 `PostBytes/PutBytes/PatchBytes/DoBytes`。
- 如果需要自定义超时、连接池或 transport，优先使用 `*WithClient` 系列 helper。
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

### 使用指南

这一节专门讲最常见的使用方式。如果你想尽快把一个 API 跑起来，直接从这里开始就够了。

#### 1. 创建应用

默认情况下，直接用 `web.New()` 即可。如果你一开始就知道要加中间件或结构化错误处理，也可以在构造时一次性声明。

```go
app := web.New(
	web.WithMiddleware(
		web.RequestID("", nil),
		web.Recover(nil),
	),
	web.WithErrorHandler(web.JSONErrorHandler(true)),
)
```

如果你希望初始化代码更集中、更声明式，优先用 `WithXxx(...)`。如果中间件是后面按条件追加的，用 `app.Use(...)` 会更自然。

#### 2. 注册路由

handler 的签名是：

```go
func(c *web.Ctx) (any, error)
```

最简单的写法是直接返回一个值，让框架帮你编码响应：

```go
app.Get("/health", func(c *web.Ctx) (any, error) {
	return map[string]string{"status": "ok"}, nil
})
```

也可以注册其他 HTTP 方法：

```go
app.Post("/users", createUser)
app.Put("/users/:id", updateUser)
app.Delete("/users/:id", deleteUser)
app.Handle("PURGE", "/cache/:key", purgeCache)
```

如果你已经有标准 `net/http` handler，也可以直接挂载：

```go
app.GetHTTP("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}))
```

#### 3. 读取路径参数、查询参数和表单

字符串场景直接用 `Param`、`Query`、`Form`。如果你想顺手完成解析和校验，优先用类型化 helper。

```go
app.Get("/users/:id", func(c *web.Ctx) (any, error) {
	id, err := c.ParamUint64("id")
	if err != nil {
		return nil, web.ErrBadRequest
	}

	verbose := c.Query("verbose") == "1"

	return map[string]any{
		"id":      id,
		"verbose": verbose,
	}, nil
})
```

表单请求示例：

```go
app.Post("/login", func(c *web.Ctx) (any, error) {
	email := c.Form("email")
	password := c.Form("password")

	if email == "" || password == "" {
		return nil, web.ErrBadRequest
	}

	return map[string]bool{"ok": true}, nil
})
```

#### 4. 解析请求体

当你希望根据请求的 `Content-Type` 自动选择解码方式时，使用 `TryParseBody`。它内置支持 JSON、GOB 和 XML。

```go
type CreateUserRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

app.Post("/users", func(c *web.Ctx) (any, error) {
	var req CreateUserRequest
	if err := c.TryParseBody(&req); err != nil {
		return nil, err
	}

	return map[string]any{
		"name": req.Name,
		"age":  req.Age,
	}, nil
})
```

如果你明确知道请求体就是 JSON，并且更关心速度，可以使用 `TryParseJSONBodyFast`。

```go
app.Post("/events", func(c *web.Ctx) (any, error) {
	var req struct {
		Type string `json:"type"`
	}

	if err := c.TryParseJSONBodyFast(&req); err != nil {
		return nil, err
	}

	return map[string]string{"accepted": req.Type}, nil
})
```

简单判断：

- 需要更严格的 JSON 语义时，用 `TryParseBody`
- 需要更快的 JSON 热路径时，用 `TryParseJSONBodyFast`

#### 5. 返回响应

框架默认的响应规则很简单：

- `return nil, nil` 写出 `204 No Content`
- `return value, nil` 写出 `200 OK`
- `c.SetStatus(code)` 可以覆写默认成功状态码
- `return nil, err` 走错误响应路径

显式设置成功状态码示例：

```go
app.Post("/users", func(c *web.Ctx) (any, error) {
	c.SetStatus(http.StatusCreated)
	return map[string]string{"result": "created"}, nil
})
```

如果你希望立即写出响应，可以使用这些 helper：

```go
app.Get("/ping", func(c *web.Ctx) (any, error) {
	return nil, c.String(http.StatusOK, "pong")
})

app.Get("/config.json", func(c *web.Ctx) (any, error) {
	return nil, c.JSON(http.StatusOK, map[string]bool{"ok": true})
})

app.Get("/logo", func(c *web.Ctx) (any, error) {
	return nil, c.Blob(http.StatusOK, "image/png", pngBytes)
})
```

一般来说，如果你的 handler 走“框架托管响应”的路线，直接返回值会更自然。只有在你需要精确控制输出时，再用 `JSON`、`String`、`Blob`、`NoContent` 这一类即时写出 helper。

#### 6. 路由分组和中间件

当前缀和中间件只想作用于某一组路由时，用 `Group` 会最清晰。

```go
api := app.Group("/api", web.Timeout(2*time.Second))
api.Use(web.AccessLog(func(c *web.Ctx, status int, d time.Duration, err error) {
	log.Printf("%s %s -> %d (%s)", c.Method(), c.Path(), status, d)
}))

api.Get("/users/:id", func(c *web.Ctx) (any, error) {
	return map[string]string{
		"id":         c.Param("id"),
		"request_id": c.RequestID(),
	}, nil
})
```

中间件执行顺序是：

- 应用级 middleware
- 父分组 middleware
- 子分组 middleware
- 路由级 middleware

几种很常用的内建 middleware 可以直接这样组合：

- `web.RequestID("", nil)` 给上下文和响应头补 request id
- `web.Recover(nil)` 把 panic 转成框架错误响应
- `web.Timeout(2 * time.Second)` 给请求加协作式超时
- `web.MaxBodyBytes(1 << 20)` 把请求体限制在 1 MiB
- `web.SecurityHeaders()` 添加一组默认安全响应头
- `web.CORSMiddleware(...)` 给命中的业务路由写出 CORS 头

#### 7. 错误处理

你可以直接返回框架内置错误：

```go
app.Get("/private", func(c *web.Ctx) (any, error) {
	if c.BearerToken() == "" {
		return nil, web.ErrUnauthorized
	}
	return map[string]bool{"ok": true}, nil
})
```

如果你希望所有 API 错误都输出统一 JSON 结构，可以安装 `JSONErrorHandler`：

```go
app.SetErrorHandler(web.JSONErrorHandler(true))
```

响应体大致会是这样：

```json
{
  "code": 401,
  "message": "UNAUTHORIZED",
  "request_id": "abc123"
}
```

如果是重定向，优先使用 `c.Redirect(...)`：

```go
app.Get("/old-home", func(c *web.Ctx) (any, error) {
	return nil, c.Redirect(http.StatusMovedPermanently, "/new-home")
})
```

#### 8. 静态文件

当应用的一部分需要暴露目录文件时，使用 `ServeFiles`：

```go
app.ServeFiles("/static/*filepath", http.Dir("./public"))
```

例如 `/static/app.css`、`/static/js/app.js` 这样的请求，会自动转发到底层文件系统，并去掉 `/static` 这一层前缀。

#### 9. 添加 CORS 和安全头

如果你的 API 要给浏览器前端调用，比较常见的组合是：

```go
app := web.New(
	web.WithCORS(web.NewCORS(web.CORSOptions{
		AllowOrigins:     []string{"https://app.example.com"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           10 * time.Minute,
	})),
)

app.Use(
	web.CORSMiddleware(web.CORSOptions{
		AllowOrigins:     []string{"https://app.example.com"},
		ExposeHeaders:    []string{"X-Request-Id"},
		AllowCredentials: true,
	}),
	web.SecurityHeaders(),
)
```

通常这两个要一起看：

- `NewCORS(...)` 负责框架自动生成的 `OPTIONS` 响应
- `CORSMiddleware(...)` 负责命中的 `GET` / `POST` / `PUT` / `PATCH` / `DELETE` 业务响应

如果你只想补常见安全头，直接上 `web.SecurityHeaders()` 就够了。需要更细的 CSP、HSTS 等策略时，再用 `SecurityHeadersWithOptions(...)`。

#### 10. 发起 HTTP 请求

这个包还带了一套轻量的客户端 helper。

简单的 JSON GET：

```go
var resp struct {
	Name string `json:"name"`
}

if err := web.Get(context.Background(), "https://api.example.com/user/1", "", &resp); err != nil {
	return err
}
```

JSON POST：

```go
payload := map[string]string{"name": "alice"}

var resp struct {
	ID uint64 `json:"id"`
}

if err := web.Post(context.Background(), "https://api.example.com/users", token, payload, &resp); err != nil {
	return err
}
```

请求体已经编码好的情况：

```go
body := []byte(`{"name":"alice"}`)
if err := web.PostBytes(context.Background(), "https://api.example.com/users", token, body, &resp); err != nil {
	return err
}
```

带重试的请求：

```go
if err := web.TryGet(context.Background(), "https://api.example.com/user/1", token, &resp, 3); err != nil {
	return err
}
```

如果你需要自定义超时、transport 或连接池，使用 `*WithClient` 版本：

```go
client := &http.Client{Timeout: 2 * time.Second}
if err := web.GetWithClient(client, context.Background(), "https://api.example.com/user/1", token, &resp); err != nil {
	return err
}
```

如果你想拿到原始响应体，而不是 JSON 解码结果：

```go
req, _ := http.NewRequest(http.MethodGet, "https://example.com/data", nil)

var raw web.RawBody
if err := web.DoReqWithClient(http.DefaultClient, req, &raw, nil); err != nil {
	return err
}
```

### API 索引

- `web.New(options ...Option) *Application`
- 路由注册：
  - `Get`, `Post`, `Put`, `Patch`, `Delete`, `Head`, `Options`, `Handle`
  - `GetHTTP`, `PostHTTP`, `PutHTTP`, `PatchHTTP`, `DeleteHTTP`, `HeadHTTP`, `OptionsHTTP`, `HandleHTTP`
- 框架组合：
  - `Use`, `Group`, `SetErrorHandler`, `RegisterReader`, `RegisterWriter`
  - options: `WithMiddleware`, `WithErrorHandler`, `WithNotFound`, `WithMethodNotAllowed`, `WithCORS`
- 服务器生命周期：
  - `ListenAndServe`, `ListenAndServeTLS`, `Shutdown`
- 辅助函数：
  - `ServeFiles`, `Redirect`, `TryParse(...)`, `TryXxx(...)`, `JSONErrorHandler`, `NewCORS`
- 上下文 (`*Ctx`) 常用方法：
  - 请求：`Method`, `Path`, `Query`, `Param`, `Body`, `ContentType`, `BearerToken`, `RequestID`
  - 解析：`TryParseBody`, `TryParseJSONBodyFast`, `TryParseParam`, `TryParseQuery`, `TryParseForm`
  - 响应：`SetHeader`, `SetCookie`, `AllowCredentials`, `JSON`, `String`, `Blob`, `NoContent`, 通过 `Accept` 进行内容协商

### API 快速参考 (CN)

| 领域 | API | 描述 |
|---|---|---|
| 应用程序 | `New()` | 创建应用程序实例 |
| 应用程序 | `New(WithMiddleware(...), WithErrorHandler(...))` | 使用构造期 options 创建应用程序 |
| 应用程序 | `Get/Post/Put/Patch/Delete/Head/Options(path, handler)` | 注册路由处理器 |
| 应用程序 | `Handle(method, path, handler)` | 为任意 HTTP 方法注册路由 |
| 应用程序 | `GetHTTP/PostHTTP/.../HandleHTTP(path, http.Handler)` | 挂载标准 `net/http` handler |
| 应用程序 | `Use(middleware...)` | 为后续注册的路由附加应用级中间件 |
| 应用程序 | `Group(prefix, middleware...)` | 创建带共享前缀和中间件的路由分组 |
| 应用程序 | `SetErrorHandler(handler)` | 安装自定义路由错误处理器 |
| 应用程序 | `SetCORS(cors)` | 为自动 `OPTIONS` 响应安装 CORS hook |
| 应用程序 | `RegisterReader(contentType, reader)` | 为指定媒体类型覆写请求解码 |
| 应用程序 | `RegisterWriter(contentType, writer)` | 为指定媒体类型覆写响应编码 |
| 应用程序 | `ServeFiles("/static/*filepath", fs)` | 使用通配路径提供静态文件服务 |
| 应用程序 | `Finalize()` | 在服务前编译路由；首次请求时会自动调用 |
| 应用程序 | `ListenAndServe(network, addr, ...opts)` | 启动 HTTP 服务器 |
| 应用程序 | `ListenAndServeTLS(network, addr, tlsConfig, ...opts)` | 启动 HTTPS 服务器 |
| 应用程序 | `Shutdown(ctx)` | 优雅关闭 |
| 上下文 | `Param(name)`, `Query(name)`, `Form(name)`, `RequestID()` | 读取路径/查询/表单值及请求 ID |
| 上下文 | `TryParseBody(v)` | 根据内容类型（JSON/GOB/XML）解析请求体 |
| 上下文 | `TryParseJSONBodyFast(v)` | 使用 pooled buffer + `json.Unmarshal` 快速解析 JSON 请求体 |
| 上下文 | `TryParseParam/Query/Form(name, &v)` | 将字符串值解析为类型化值 |
| 上下文 | `SetHeader`, `SetCookie`, `SetContentType`, `SetStatus` | 写入响应头并覆写默认成功状态码 |
| 上下文 | `JSON`, `String`, `Blob`, `NoContent` | 用于显式写出的即时响应 helper |
| 上下文 | `Request()`, `ResponseWriter()`, `Context()` | 访问原始 HTTP 对象 |
| 中间件 | `RequestID`, `Recover`, `RecoverWithOptions`, `Timeout`, `AccessLog`, `AccessLogWithOptions` | 核心内建中间件 |
| 中间件 | `MaxBodyBytes(limit)` | 限制请求体大小 |
| 中间件 | `SecurityHeaders()` / `SecurityHeadersWithOptions(...)` | 添加常见安全响应头 |
| 中间件 | `CORSMiddleware(CORSOptions)` | 为命中的业务路由写出 CORS 头 |
| CORS helper | `NewCORS(CORSOptions)` | 为自动 `OPTIONS` 处理创建 CORS hook |
| 客户端 | `Get/Post/Put/Patch/Delete/Do` | 使用 `http.DefaultClient` 的 HTTP 辅助函数 |
| 客户端 | `GetWithClient/PostWithClient/PutWithClient/PatchWithClient/DeleteWithClient/DoWithClient` | 显式传入 `*http.Client` 的 HTTP 辅助函数 |
| 客户端 | `DoReq/DoReqWithClient` | 执行已构造请求，并解码 JSON 或 `RawBody` 响应体 |
| 客户端 | `PostBytes/PutBytes/PatchBytes/DoBytes` | 发送预编码请求体，绕过 JSON 编码 |
| 客户端 | `PostBytesWithClient/PutBytesWithClient/PatchBytesWithClient/DoBytesWithClient` | 显式传入 `*http.Client` 的预编码请求体辅助函数 |
| 客户端 | `TryGet/TryPost/TryPut/TryPatch/TryDelete/TryDo` | 带有重试循环的 HTTP 辅助函数 |
| 客户端 | `TryGetWithClient/TryPostWithClient/TryPutWithClient/TryPatchWithClient/TryDeleteWithClient/TryDoWithClient` | 显式传入 `*http.Client` 的重试辅助函数 |
| 客户端 | `TryPostBytes/TryPutBytes/TryPatchBytes/TryDoBytes` | 预编码请求体的重试辅助函数 |
| 客户端 | `TryPostBytesWithClient/TryPutBytesWithClient/TryPatchBytesWithClient/TryDoBytesWithClient` | 显式传入 `*http.Client` 的预编码请求体重试辅助函数 |
| 错误 | `NewErr(code, msg)` | 带有 HTTP 状态码的错误 |
| 错误 | `Redirect(url, code)` | 从处理器返回重定向响应 |
| 错误 | `JSONErrorHandler(includeRequestID)` | 输出结构化 JSON API 错误 |

### 响应行为

- 处理器返回值控制响应：
  - `(nil, nil)` -> `204 No Content`
  - `(value, nil)` -> `200 OK`
  - 调用 `c.SetStatus(code)` 可以显式覆写默认成功状态码
  - `(_, err)` -> 状态码来自框架错误类型，响应体包含 `err.Error()`
- 响应格式通过请求的 `Accept` 头部选择：
  - `application/json`
  - `application/x-gob`
  - `application/xml`
  - `application/octet-stream`
  - `application/x-avro`

### 路由行为

- 同一层路由树允许静态子段、参数子段和末尾 catch-all 共存。
- 命中优先级固定，且不依赖注册顺序：
  - `static > param > catch-all`
- catch-all 仍然只能出现在路径末尾。
- 非法 wildcard 组合在注册时仍会被拒绝。

这样就可以直接表达常见 REST 路由集，而不需要在 handler 里做 catch-all 分发：

```go
app.Get("/organizations/:id/devices/bulk/disable", bulkDisable)
app.Get("/organizations/:id/devices/provision", provision)
app.Get("/organizations/:id/devices/config/rollout", configRollout)
app.Get("/organizations/:id/devices/:device_id", showDevice)
```

对于上面的路由：

- `GET /organizations/1/devices/provision` 会命中静态路由。
- `GET /organizations/1/devices/42` 会命中参数路由。
- 无论先注册 `:device_id` 还是先注册 `provision`，结果都一致。

### 现代框架能力

- 中间件和路由分组在“注册期”完成，不在请求期动态拼装：
  - `app.Use(...)`
  - `app.Group("/api", ...)`
  - 分组级 `Use(...)`
- 内建中间件为显式 opt-in：
  - `RequestID`
  - `Recover`
  - `RecoverWithOptions`
  - `Timeout`
  - `AccessLog`
  - `AccessLogWithOptions`
  - `MaxBodyBytes`
  - `SecurityHeaders`
  - `SecurityHeadersWithOptions`
  - `CORSMiddleware`
- 结构化 API 错误通过 `SetErrorHandler(JSONErrorHandler(...))` 显式启用
- 自动 `OPTIONS` 的 CORS 响应通过 `SetCORS(NewCORS(...))` 显式启用
- `Reader/Writer` 覆写按媒体类型注册；未注册时不会影响默认热路径
- 构造期 options 让应用初始化更声明式，同时不增加请求期成本。
- 已有 `net/http` handler 可以通过 `HandleHTTP`/`GetHTTP` 直接挂载。

```go
app := web.New(
	web.WithMiddleware(web.RequestID("", nil), web.Recover(nil)),
	web.WithErrorHandler(web.JSONErrorHandler(true)),
)

api := app.Group("/api", web.Timeout(2*time.Second))
api.Get("/users/:id", func(c *web.Ctx) (any, error) {
	return map[string]string{
		"id":         c.Param("id"),
		"request_id": c.RequestID(),
	}, nil
})

api.GetHTTP("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}))
api.Get("/avatar/:id", func(c *web.Ctx) (any, error) {
	return nil, c.Blob(http.StatusOK, "image/png", []byte{0x89, 'P', 'N', 'G'})
})
```

如果你需要更细的控制，可以使用带 options 的中间件变体：

```go
app.Use(
	web.RecoverWithOptions(web.RecoverOptions{
		DefaultStatus: http.StatusServiceUnavailable,
		DefaultBody:   "UNAVAILABLE",
	}),
	web.AccessLogWithOptions(web.AccessLogOptions{
		Log: func(c *web.Ctx, entry web.AccessLogEntry) {
			// 在这里接入路由级访问日志
		},
	}),
)
```

### 兼容性 / 破坏性变更

- `Try*` 重试语义更新：
  - `retry <= 0` 现在仍执行一次请求尝试。
  - 对于 `ErrUnauthorized`、`ErrForbidden` 和 `ErrBadRequest`（包括包装后的），重试循环会提早停止。
- `TryDo` 现在支持跨重试的请求体安全重放（请求体会被缓冲一次并在每次尝试时重新创建）。
- 新增原始请求体辅助函数：
  - `PostBytes`, `PutBytes`, `PatchBytes`, `DoBytes`
  - `TryPostBytes`, `TryPutBytes`, `TryPatchBytes`, `TryDoBytes`
  - 默认请求头为 `Content-Type: application/octet-stream` 和 `Accept: application/json`
- 新增显式 client 辅助函数：
  - `DoReqWithClient`, `DoWithClient`, `DoBytesWithClient`
  - `Get/Post/Put/Patch/Delete` 以及对应重试版本都提供 `*WithClient` 变体
  - 需要 transport 级性能调优时应优先使用
- 新增原始响应体快速路径：
  - `DoReq` / `DoReqWithClient` 可识别 `*web.RawBody`
  - `[]byte`、`json.RawMessage` 等既有 JSON 目标类型保持原有 JSON 语义
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

### JSON 请求体快路径

当请求体是 JSON，且你不需要“拒绝未知字段”这一语义时，可以使用 `TryParseJSONBodyFast`。

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

### 客户端原始响应体

当你希望拿到原始响应 payload，而不是做 JSON 解码时，可以把 `DoReqWithClient` 的目标参数写成 `*web.RawBody`。

```go
req, _ := http.NewRequest(http.MethodGet, "https://example.com/data", nil)

var raw web.RawBody
if err := web.DoReqWithClient(client, req, &raw, nil); err != nil {
	panic(err)
}
```

### 注意事项

- 参数/通配路由的最佳性能是在参数被池化时实现的（`Application` 中已使用）。
- 对于二进制/Avro 响应，首选返回 `[]byte` 或实现 `web.AvroMarshaler` 以避免额外的编码开销。
- `TryParseBody` 目前仅支持 JSON/GOB/XML。

### 致谢

感谢所有开源项目，我从中受益匪浅。

特别感谢：

- [httprouter](https://github.com/julienschmidt/httprouter): 一个高性能的 HTTP 路由，启发了本项目中的路由逻辑。
- [web](https://github.com/hoisie/web): 一个轻量级的 Web 框架，提供了关于高效服务器设计的见解。
