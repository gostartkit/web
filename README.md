# Web.go

### Route
```go
package route

import (
	"github.com/gostartkit/auth/controller"
	"github.com/gostartkit/auth/middleware"
	"github.com/webpkg/web"
)

func userRoute(app *web.Application, prefix string) {

	user := controller.CreateUserController()

	app.Get(prefix+"/user/", middleware.Chain(user.Index, "user.all"))
	app.Post(prefix+"/user/", middleware.Chain(user.Create, "user.edit"))
	app.Get(prefix+"/user/:id", middleware.Chain(user.Detail, "user.all"))
	app.Put(prefix+"/user/:id", middleware.Chain(user.Update, "user.edit"))
	app.Patch(prefix+"/user/:id/status/", middleware.Chain(user.UpdateStatus, "user.edit"))
	app.Delete(prefix+"/user/:id", middleware.Chain(user.Destroy, "user.edit"))
}

```

### Controller
```go
package controller

import (
	"sync"

	"github.com/gostartkit/auth/model"
	"github.com/gostartkit/auth/proxy"
	"github.com/gostartkit/auth/validator"
	"github.com/webpkg/web"
)

var (
	_userController     *UserController
	_onceUserController sync.Once
)

// CreateUserController return UserController
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
func (c *UserController) Index(ctx *web.Context) (web.Data, error) {
	var (
		page     int
		pageSize int
	)

	filter := ctx.Query("$filter")
	orderBy := ctx.Query("$orderBy")
	ctx.TryParseQuery("$page", &page)
	ctx.TryParseQuery("$pageSize", &pageSize)

	return proxy.GetUsers(filter, orderBy, page, pageSize)
}

// Create create user
func (c *UserController) Create(ctx *web.Context) (web.Data, error) {
	user := model.CreateUser()

	if err := ctx.TryParseBody(user); err != nil {
		return nil, err
	}

	if err := validator.CreateUser(user); err != nil {
		return nil, err
	}

	return proxy.CreateUser(user)
}

// Detail get user
func (c *UserController) Detail(ctx *web.Context) (web.Data, error) {
	var (
		id uint64
	)

	if err := ctx.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	return proxy.GetUser(id)
}

// Update update user by id
func (c *UserController) Update(ctx *web.Context) (web.Data, error) {
	user := model.CreateUser()

	if err := ctx.TryParseBody(user); err != nil {
		return nil, err
	}

	if err := ctx.TryParseParam("id", &user.ID); err != nil {
		return nil, err
	}

	if err := validator.UpdateUser(user); err != nil {
		return nil, err
	}

	return proxy.UpdateUser(user)
}

// UpdateStatus update user.Status by id
func (c *UserController) UpdateStatus(ctx *web.Context) (web.Data, error) {
	user := model.CreateUser()

	if err := ctx.TryParseBody(user); err != nil {
		return nil, err
	}

	if err := ctx.TryParseParam("id", &user.ID); err != nil {
		return nil, err
	}

	if err := validator.UpdateUserStatus(user); err != nil {
		return nil, err
	}

	return proxy.UpdateUserStatus(user)
}

// Destroy delete user
func (c *UserController) Destroy(ctx *web.Context) (web.Data, error) {
	var (
		id uint64
	)

	if err := ctx.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	return proxy.DestroyUserSoft(id)
}

```
### Thanks
Thanks for all open source projects, I learned a lot from them.

Special thanks to these two projects：

https://github.com/julienschmidt/httprouter

https://github.com/hoisie/web