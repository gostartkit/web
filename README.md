# Web.go The library for web

### Route
```go
package route

import (
	"github.com/gostartkit/auth/controller"
	"github.com/gostartkit/auth/middleware"
	"pkg.gostartkit.com/web"
)

func userRoute(app *web.Application, prefix string) {

	user := controller.CreateUserController()

	app.Get(prefix+"/user/", middleware.Chain(user.Index, "user.all"))
	app.Get(prefix+"/user/:id", middleware.Chain(user.Detail, "user.all"))
	app.Post(prefix+"/apply/user/id/", middleware.Chain(user.CreateID, "user.edit"))
	app.Post(prefix+"/user/", middleware.Chain(user.Create, "user.edit"))
	app.Put(prefix+"/user/:id", middleware.Chain(user.Update, "user.edit"))
	app.Patch(prefix+"/user/:id", middleware.Chain(user.UpdatePartial, "user.edit"))
	app.Patch(prefix+"/user/:id/status/", middleware.Chain(user.UpdateStatus, "user.edit"))
	app.Delete(prefix+"/user/:id", middleware.Chain(user.Destroy, "user.edit"))
	app.Get(prefix+"/user/:id/application/", middleware.Chain(user.Applications, "user.all"))
	app.Post(prefix+"/user/:id/application/", middleware.Chain(user.LinkApplications, "user.edit"))
	app.Delete(prefix+"/user/:id/application/", middleware.Chain(user.UnLinkApplications, "user.edit"))
	app.Get(prefix+"/user/:id/application/:applicationID", middleware.Chain(user.Application, "user.all"))
	app.Put(prefix+"/user/:id/application/:applicationID", middleware.Chain(user.UpdateApplication, "user.edit"))
	app.Get(prefix+"/user/:id/role/", middleware.Chain(user.Roles, "user.all"))
	app.Post(prefix+"/user/:id/role/", middleware.Chain(user.LinkRoles, "user.edit"))
	app.Delete(prefix+"/user/:id/role/", middleware.Chain(user.UnLinkRoles, "user.edit"))
}
```

### 
```go
package controller

import (
	"strings"
	"sync"

	"github.com/gostartkit/auth/model"
	"github.com/gostartkit/auth/proxy"
	"github.com/gostartkit/auth/validator"
	"pkg.gostartkit.com/web"
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
func (o *UserController) Index(c *web.Ctx) (web.Any, error) {

	var (
		page     int
		pageSize int
	)

	filter := c.Query(web.QueryFilter)
	orderBy := c.Query(web.QueryOrderBy)
	c.TryParseQuery(web.QueryPage, &page)
	c.TryParseQuery(web.QueryPageSize, &pageSize)

	return proxy.GetUsers(filter, orderBy, page, pageSize)
}

// Detail get user
func (o *UserController) Detail(c *web.Ctx) (web.Any, error) {

	var id uint64

	if err := c.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	if id == 0 {
		return nil, validator.CreateInvalidError("id")
	}

	return proxy.GetUser(id)
}

// CreateID create user.ID
func (o *UserController) CreateID(c *web.Ctx) (web.Any, error) {
	return proxy.CreateUserID()
}

// Create create user
func (o *UserController) Create(c *web.Ctx) (web.Any, error) {

	user := model.NewUser()

	if err := c.TryParseBody(user); err != nil {
		return nil, err
	}

	if err := validator.CreateUser(user); err != nil {
		return nil, err
	}

	if _, err := proxy.CreateUser(user); err != nil {
		return nil, err
	}

	return user.ID, nil
}

// Update update user
func (o *UserController) Update(c *web.Ctx) (web.Any, error) {

	user := model.NewUser()

	if err := c.TryParseBody(user); err != nil {
		return nil, err
	}

	if err := c.TryParseParam("id", &user.ID); err != nil {
		return nil, err
	}

	if err := validator.UpdateUser(user); err != nil {
		return nil, err
	}

	return proxy.UpdateUser(user)
}

// UpdatePartial update user
func (o *UserController) UpdatePartial(c *web.Ctx) (web.Any, error) {

	attrs := strings.Split(c.Get(web.HeaderAttrs), ",")

	if len(attrs) == 0 {
		return nil, validator.CreateRequiredError(web.HeaderAttrs)
	}

	user := model.NewUser()

	if err := c.TryParseBody(user); err != nil {
		return nil, err
	}

	if err := c.TryParseParam("id", &user.ID); err != nil {
		return nil, err
	}

	if err := validator.UpdateUserPartial(user, attrs...); err != nil {
		return nil, err
	}

	return proxy.UpdateUserPartial(user, attrs...)
}

// UpdateStatus update user.Status
func (o *UserController) UpdateStatus(c *web.Ctx) (web.Any, error) {

	user := model.NewUser()

	if err := c.TryParseBody(user); err != nil {
		return nil, err
	}

	if err := c.TryParseParam("id", &user.ID); err != nil {
		return nil, err
	}

	if err := validator.UpdateUserStatus(user); err != nil {
		return nil, err
	}

	return proxy.UpdateUserStatus(user)
}

// Destroy delete user
func (o *UserController) Destroy(c *web.Ctx) (web.Any, error) {

	var id uint64

	if err := c.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	if id == 0 {
		return nil, validator.CreateInvalidError("id")
	}

	return proxy.DestroyUserSoft(id)
}

// Applications return *model.ApplicationCollection, error
func (o *UserController) Applications(c *web.Ctx) (web.Any, error) {

	var (
		id       uint64
		page     int
		pageSize int
	)

	if err := c.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	filter := c.Query(web.QueryFilter)
	orderBy := c.Query(web.QueryOrderBy)
	c.TryParseQuery(web.QueryPage, &page)
	c.TryParseQuery(web.QueryPageSize, &pageSize)

	return proxy.GetApplicationsByUserID(id, filter, orderBy, page, pageSize)
}

// LinkApplications return rowsAffected int64, error
func (o *UserController) LinkApplications(c *web.Ctx) (web.Any, error) {

	var (
		id            uint64
		applicationID []uint64
	)

	if err := c.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	if err := c.TryParseBody(&applicationID); err != nil {
		return nil, err
	}

	return proxy.LinkUserApplications(id, applicationID...)
}

// UnLinkApplications return rowsAffected int64, error
func (o *UserController) UnLinkApplications(c *web.Ctx) (web.Any, error) {

	var (
		id            uint64
		applicationID []uint64
	)

	if err := c.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	if err := c.TryParseBody(&applicationID); err != nil {
		return nil, err
	}

	return proxy.UnLinkUserApplications(id, applicationID...)
}

// Application return *model.ApplicationUser, error
func (o *UserController) Application(c *web.Ctx) (web.Any, error) {

	var (
		id            uint64
		applicationID uint64
	)

	if err := c.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	if err := c.TryParseParam("applicationID", &applicationID); err != nil {
		return nil, err
	}

	return proxy.GetUserApplication(id, applicationID)
}

// UpdateApplication return rowsAffected int64, error
func (o *UserController) UpdateApplication(c *web.Ctx) (web.Any, error) {

	applicationUser := model.NewApplicationUser()

	if err := c.TryParseBody(applicationUser); err != nil {
		return nil, err
	}

	if err := c.TryParseParam("applicationID", &applicationUser.ApplicationID); err != nil {
		return nil, err
	}

	if err := c.TryParseParam("id", &applicationUser.UserID); err != nil {
		return nil, err
	}

	if err := validator.UpdateUserApplication(applicationUser); err != nil {
		return nil, err
	}

	return proxy.UpdateUserApplication(applicationUser)
}

// Roles return *model.RoleCollection, error
func (o *UserController) Roles(c *web.Ctx) (web.Any, error) {

	var (
		id       uint64
		page     int
		pageSize int
	)

	if err := c.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	filter := c.Query(web.QueryFilter)
	orderBy := c.Query(web.QueryOrderBy)
	c.TryParseQuery(web.QueryPage, &page)
	c.TryParseQuery(web.QueryPageSize, &pageSize)

	return proxy.GetRolesByUserID(id, filter, orderBy, page, pageSize)
}

// LinkRoles return rowsAffected int64, error
func (o *UserController) LinkRoles(c *web.Ctx) (web.Any, error) {

	var (
		id     uint64
		roleID []uint64
	)

	if err := c.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	if err := c.TryParseBody(&roleID); err != nil {
		return nil, err
	}

	return proxy.LinkUserRoles(id, roleID...)
}

// UnLinkRoles return rowsAffected int64, error
func (o *UserController) UnLinkRoles(c *web.Ctx) (web.Any, error) {

	var (
		id     uint64
		roleID []uint64
	)

	if err := c.TryParseParam("id", &id); err != nil {
		return nil, err
	}

	if err := c.TryParseBody(&roleID); err != nil {
		return nil, err
	}

	return proxy.UnLinkUserRoles(id, roleID...)
}
```
### Thanks
Thanks for all open source projects, I learned a lot from them.

Special thanks to these two projects：

https://github.com/julienschmidt/httprouter

https://github.com/hoisie/web