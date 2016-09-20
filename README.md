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

To run the application, put the code in a file called hello.go and run:

    go run hello.go
    
You can point your browser to http://localhost:8080/hello/world . 

## Reference Links

https://github.com/julienschmidt/httprouter

https://github.com/hoisie/web