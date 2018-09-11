package main

import "github.com/webpkg/web"

func main() {
	app := web.CreateApplication()

	app.Use(func(ctx *web.Context) {

	})

	app.Listen(":3000")
}
