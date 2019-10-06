package main

import "github.com/webpkg/web"

func main() {
	app := web.Create()

	app.Use(func(ctx *web.Context) {
		ctx.Response.Write("001")
	})

	app.Resource("user/", func(ctx *web.Context) {
		ctx.Response.Write("002")
	})

	app.Use(func(ctx *web.Context) {
		ctx.Response.Write("003")
	})

	app.ListenAndServe(":http")
}
