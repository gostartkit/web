# Web.go

## Graceful Shutdown

```bash
kill -2 $PID
```

```go
signal.Notify(sigint, os.Interrupt) // kill -2 pid
signal.Notify(sigint, syscall.SIGTERM) // kill pid
```
### Route
```go
package route

import (
	"sync"

	"github.com/webpkg/api/controller"
	"github.com/webpkg/api/middleware"
	"github.com/webpkg/web"
)

var (
	_once sync.Once
)

// Init config
func Init(app *web.Application) {

	_once.Do(func() {
		user := controller.CreateUserController()
		app.Get("/user/", middleware.Auth(user.Index, "user.all"))
		app.Post("/user/", middleware.Auth(user.Create, "user.edit"))
		app.Post("/user/:id", middleware.Auth(user.LinkRoles, "user.edit", "role.edit"))
		app.Get("/user/:id", middleware.Auth(user.Detail, "user.all|user.self"))
		app.Get("/user/:id/role/", middleware.Auth(user.Roles, "role.all"))
		app.Patch("/user/:id", middleware.Auth(user.Update, "user.edit"))
		app.Put("/user/:id", middleware.Auth(user.Update, "user.edit"))
		app.Delete("/user/:id", middleware.Auth(user.Destroy, "user.edit"))
		app.Delete("/user/:id/role/", middleware.Auth(user.UnLinkRoles, "user.edit", "role.edit"))
	})
}

```

### Controller
```go
package controller

import (
	"sync"

	"github.com/webpkg/api/helper"
	"github.com/webpkg/api/model"
	"github.com/webpkg/api/proxy"
	"github.com/webpkg/api/rbac"
	"github.com/webpkg/api/validator"
	"github.com/webpkg/web"
)

var (
	_userController     *UserController
	_onceUserController sync.Once
)

// CreateUserController return *UserController
func CreateUserController() *UserController {

	_onceUserController.Do(func() {
		_userController = &UserController{}
	})

	return _userController
}

// UserController struct
type UserController struct {
}

// Index get users
func (c *UserController) Index(ctx *web.Context) {
	var (
		page     int
		pageSize int
	)

	key := ctx.Query("key")
	ctx.TryParseQuery("page", &page)
	ctx.TryParseQuery("pagesize", &pageSize)

	ctx.AbortIf(proxy.GetUsersByKey(key, page, pageSize))
}

// Create create user
func (c *UserController) Create(ctx *web.Context) {
	user := model.CreateUser()
	ctx.Parse(user)
	ctx.Abort(validator.CreateUser(user))

	ctx.Abort(proxy.CreateUser(user))
}

// Detail get user detail by id
func (c *UserController) Detail(ctx *web.Context) {
	var id uint64

	ctx.ParseParam("id", &id)

	ctx.AbortIf(proxy.GetUser(id))
}

// Update update user by id
func (c *UserController) Update(ctx *web.Context) {
	user := model.CreateUser()
	ctx.Parse(user)
	ctx.Abort(validator.UpdateUser(user))

	ctx.AbortIf(proxy.UpdateUser(user))
}

// Destroy delete user by id
func (c *UserController) Destroy(ctx *web.Context) {
	var id uint64

	ctx.ParseParam("id", &id)

	ctx.AbortIf(proxy.DestroyUserSoft(id))
}

```
### Thanks
Thanks for all open source projects, I learned a lot from them.

Special thanks to these two projects：

https://github.com/julienschmidt/httprouter

https://github.com/hoisie/web