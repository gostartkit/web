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

	"github.com/webpkg/web"
	"github.com/webpkg/api/controller"
)

var (
	_once sync.Once
)

// Init config
func Init(app *web.Application) {

	_once.Do(func() {
		app.Resource("/user/", controller.CreateUserController())
	})
}

```

### Controller
```go
package controller

import (
	"sync"

	"github.com/webpkg/web"
)

var (
	_userController     web.Controller
	_onceUserController sync.Once
)

// CreateUserController return web.Controller
func CreateUserController() web.Controller {

	_onceUserController.Do(func() {
		_userController = &userController{}
	})

	return _userController
}

// userController struct
type userController struct {
}

// Index get users
func (uc *userController) Index(ctx *web.Context) {
	ctx.WriteString("user.index")
}

// Create create user
func (uc *userController) Create(ctx *web.Context) {
	ctx.WriteString("user.create")
}

// Detail get user detail by id
func (uc *userController) Detail(ctx *web.Context) {
	ctx.WriteString("user.detail")
}

// Update update user by id
func (uc *userController) Update(ctx *web.Context) {
	ctx.WriteString("user.update")
}

// Destroy delete user by id
func (uc *userController) Destroy(ctx *web.Context) {
	ctx.WriteString("user.destroy")
}

```
### Thanks
Thanks for all open source projects, I learned a lot from them.

Special thanks to these two projects：

https://github.com/julienschmidt/httprouter

https://github.com/hoisie/web