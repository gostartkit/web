# web
a MVC Web Application Framework for golang

## Installation

Make sure you have the a working Go environment. See the [install instructions](http://golang.org/doc/install.html). web.go targets the Go `release` branch.

To install web.go, simply run:

    go get github.com/afxcn/web

To compile it from source:

    git clone git://github.com/afxcn/web.git
    cd web && go build

## Example

Hello world:

```go
package main

import (
	"fmt"

	"github.com/afxcn/web"
)

func Index(ctx *web.Context) {
	fmt.Fprint(ctx.ResponseWriter, "Welcome!\n")
}

func Hello(ctx *web.Context) {
	ctx.WriteString("Hello " + ctx.Val("name") + "\n")
}

func Hello2(ctx *web.Context) {
	fmt.Fprint(ctx.ResponseWriter, "hello again with post.?\n")
}

func main() {
	s := web.CreateServer()
	s.Get("/", Index)
	s.Get("/hello/:name", Hello)
	s.Post("/hello/", Hello2)
	s.Run("127.0.0.1:8080")
}
```

BasicAuth:

```go
package main

import (
	"io"
	"net/http"

	"github.com/afxcn/web"
)

func main() {
	server := web.CreateServer()
	server.Get("/", Index)
	server.Get("/auth/", BasicAuth(AuthSuccess, "user", "pass"))

	server.Run("127.0.0.1:8080")
}

func Index(ctx *web.Context) {
	ctx.SetContentType("html")
	ctx.SetHeader("Access-Control-Allow-Origin", "/", true)
	ctx.SetCookie("user_id", "1728727338923", 3600)

	io.WriteString(ctx.ResponseWriter, "Webcome to web api system.\n")
}

func AuthSuccess(ctx *web.Context) {
	ctx.WriteString("Auth Success\n")
}

func BasicAuth(handler web.Handler, requiredUser, requiredPassword string) web.Handler {
	return func(ctx *web.Context) {
		user, password, hasAuth := ctx.Request.BasicAuth()

		if hasAuth && user == requiredUser && password == requiredPassword {
			handler(ctx)
		} else {
			ctx.SetHeader("WWW-Authenticate", "Basic realm=Restricted", true)
			http.Error(ctx.ResponseWriter, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}
```

To run the application, put the code in a file called hello.go and run:

    go run hello.go
    
You can point your browser to http://localhost:8080/hello/world . 

## Reference Links

https://github.com/julienschmidt/httprouter

https://github.com/hoisie/web