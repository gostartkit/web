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
	"github.com/webpkg/web"
)

var (
	_once sync.Once
)

// Init config
func Init(app *web.Application) {

	_once.Do(func() {
		user := controller.CreateUserController()
		app.Get("/user/", user.Index)
		app.Post("/user/", user.Create)
		app.Get("/user/:id", user.Detail)
		app.Patch("/user/:id", user.Update)
		app.Put("/user/:id", user.Update)
		app.Delete("/user/:id", user.Destroy)
	})
}

```

### Controller
```go
package controller

import (
	"log"
	"sync"

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
func (uc *UserController) Index(ctx *web.Context) {
	ctx.WriteString("user.index")
}

// Create create user
func (uc *UserController) Create(ctx *web.Context) {

	name := ctx.Form("name")
	log.Printf("%s", name)
	ctx.WriteString(name)
}

// Detail get user detail by id
func (uc *UserController) Detail(ctx *web.Context) {
	ctx.WriteString("user.detail")
}

// Update update user by id
func (uc *UserController) Update(ctx *web.Context) {
	ctx.WriteString("user.update")
}

// Destroy delete user by id
func (uc *UserController) Destroy(ctx *web.Context) {
	ctx.WriteString("user.destroy")
}

```
### Thanks
Thanks for all open source projects, I learned a lot from them.

Special thanks to these two projects：

https://github.com/julienschmidt/httprouter

https://github.com/hoisie/web