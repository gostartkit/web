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

// CreateUserController return web.Controller
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

	users, err := proxy.GetUsersByKey(key, page, pageSize)
	ctx.AbortError(0, validator.Error(err))

	ctx.WriteSuccess(0, users)
}

// Create create user
func (c *UserController) Create(ctx *web.Context) {
	user := model.CreateUser()
	ctx.Parse(&user)
	ctx.AbortError(0, validator.CreateUser(&user))

	user.Password = helper.Hash(user.Password)

	userID, err := proxy.CreateUser(&user)
	ctx.AbortError(0, validator.Error(err))

	ctx.WriteSuccess(0, userID)
}

// Detail get user detail by id
func (c *UserController) Detail(ctx *web.Context) {
	var id uint64

	ctx.ParseParam("id", &id)
	user, err := proxy.GetUser(id)
	ctx.AbortError(0, validator.Error(err))

	ctx.WriteSuccess(0, user)
}

// Update update user by id
func (c *UserController) Update(ctx *web.Context) {
	user := model.CreateUser()
	ctx.Parse(&user)
	ctx.AbortError(0, validator.UpdateUser(&user))

	var (
		rowsAffected int64
		err          error
	)

	if user.Password == "" {
		rowsAffected, err = proxy.UpdateUser(&user)
	} else {
		user.Password = helper.Hash(user.Password)
		rowsAffected, err = proxy.UpdateUserWithPassword(&user)
	}

	ctx.AbortError(0, validator.Error(err))

	ctx.WriteSuccess(0, rowsAffected)
}

// Destroy delete user by id
func (c *UserController) Destroy(ctx *web.Context) {
	var id uint64

	ctx.ParseParam("id", &id)
	rowsAffected, err := proxy.DestroyUserSoft(id)
	ctx.AbortError(0, validator.Error(err))

	ctx.WriteSuccess(0, rowsAffected)
}

// CurrentUserRight get user rights
func (c *UserController) CurrentUserRight(ctx *web.Context) {
	rights, err := rbac.GetUserRights(ctx.UserID)
	ctx.AbortError(0, validator.Error(err))
	ctx.WriteSuccess(0, rights)
}

// Right get user rights
func (c *UserController) Right(ctx *web.Context) {
	var id uint64
	ctx.ParseParam("id", &id)

	rights, err := rbac.GetUserRights(id)
	ctx.AbortError(0, validator.Error(err))
	ctx.WriteSuccess(0, rights)
}

// UpdateRight get user rights
func (c *UserController) UpdateRight(ctx *web.Context) {
	var id uint64
	ctx.ParseParam("id", &id)

	var rights []string
	ctx.Parse(&rights)

	right := rbac.ConvertToRight(rights)

	rowsAffected, err := proxy.UpdateUserRight(id, right)
	ctx.AbortError(0, validator.Error(err))

	ctx.WriteSuccess(0, rowsAffected)
}

// Roles get roles by userID
func (c *UserController) Roles(ctx *web.Context) {
	var (
		id       uint64
		page     int
		pageSize int
	)

	ctx.ParseParam("id", &id)
	ctx.TryParseQuery("page", &page)
	ctx.TryParseQuery("pagesize", &pageSize)

	roles, err := proxy.GetRolesByUserID(id, page, pageSize)
	ctx.AbortError(0, validator.Error(err))

	ctx.WriteSuccess(0, roles)
}

// LinkRoles link roles to user
func (c *UserController) LinkRoles(ctx *web.Context) {
	var (
		id      uint64
		rolesID []uint64
	)
	ctx.ParseParam("id", &id)
	ctx.Parse(&rolesID)

	rowsAffected, err := proxy.LinkUserRoles(id, rolesID)
	ctx.AbortError(0, validator.Error(err))

	ctx.WriteSuccess(0, rowsAffected)
}

// UnLinkRoles unlink user and roles
func (c *UserController) UnLinkRoles(ctx *web.Context) {
	var (
		id      uint64
		rolesID []uint64
	)
	ctx.ParseParam("id", &id)
	ctx.Parse(&rolesID)

	rowsAffected, err := proxy.UnLinkUserRoles(id, rolesID)
	ctx.AbortError(0, validator.Error(err))

	ctx.WriteSuccess(0, rowsAffected)
}

```
### Thanks
Thanks for all open source projects, I learned a lot from them.

Special thanks to these two projects：

https://github.com/julienschmidt/httprouter

https://github.com/hoisie/web