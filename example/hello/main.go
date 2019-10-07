package main

import (
	"log"

	"github.com/webpkg/web"
)

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

	app.Get("/user/:name", func(ctx *web.Context) {
		ctx.Response.Write("Sander")
	})

	// m := autocert.Manager{
	// 	Prompt:     autocert.AcceptTOS,
	// 	Cache:      autocert.DirCache("certs"),
	// 	HostPolicy: autocert.HostWhitelist("ip.onlineplaytime.com"),
	// }

	// tlsConfig := &tls.Config{
	// 	GetCertificate: m.GetCertificate,
	// }

	log.Fatal(app.ListenAndServe(":http"))
}
