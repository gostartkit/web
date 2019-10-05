package main

import "github.com/webpkg/web"

func main() {
	app := web.Singleton()

	app.Use(func(ctx *web.Context) {

	})

	app.Listen(":http")
}
