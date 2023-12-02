# Web.go The library for web

### Route
```go
package route

import (
	"gostartkit.com/go/auth/config"
	"gostartkit.com/go/auth/controller"
	"gostartkit.com/go/auth/middleware"
	"pkg.gostartkit.com/web"
)

func userRoute(app *web.Application, prefix string) {

	user := controller.CreateUserController()

	app.Get(prefix+"/user/", middleware.Chain(user.Index, config.Read|config.ReadUser))
	app.Get(prefix+"/user/:id", middleware.Chain(user.Detail, config.Read|config.ReadUser))
	app.Post(prefix+"/apply/user/id/", middleware.Chain(user.CreateID, config.Write|config.WriteUser))
	app.Post(prefix+"/user/", middleware.Chain(user.Create, config.Write|config.WriteUser))
	app.Put(prefix+"/user/:id", middleware.Chain(user.Update, config.Write|config.WriteUser))
	app.Patch(prefix+"/user/:id", middleware.Chain(user.Patch, config.Write|config.WriteUser))
	app.Patch(prefix+"/user/:id/status/", middleware.Chain(user.UpdateStatus, config.Write|config.WriteUser))
	app.Delete(prefix+"/user/:id", middleware.Chain(user.Destroy, config.Write|config.WriteUser))
	app.Get(prefix+"/user/:id/application/", middleware.Chain(user.Applications, config.Read|config.ReadUser))
	app.Post(prefix+"/user/:id/application/", middleware.Chain(user.LinkApplications, config.Write|config.WriteUser))
	app.Delete(prefix+"/user/:id/application/", middleware.Chain(user.UnLinkApplications, config.Write|config.WriteUser))
	app.Get(prefix+"/user/:id/application/:applicationID", middleware.Chain(user.Application, config.Read|config.ReadUser))
	app.Put(prefix+"/user/:id/application/:applicationID", middleware.Chain(user.UpdateApplication, config.Write|config.WriteUser))
	app.Get(prefix+"/user/:id/role/", middleware.Chain(user.Roles, config.Read|config.ReadUser))
	app.Post(prefix+"/user/:id/role/", middleware.Chain(user.LinkRoles, config.Write|config.WriteUser))
	app.Delete(prefix+"/user/:id/role/", middleware.Chain(user.UnLinkRoles, config.Write|config.WriteUser))
}

```

### Controller
```go
package controller

import (
	"sync"

	"gostartkit.com/go/auth/model"
	"gostartkit.com/go/auth/proxy"
	"gostartkit.com/go/auth/validator"
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

	filter := c.QueryFilter()
	orderBy := c.QueryOrderBy()
	page := c.QueryPage(_defaultPage)
	pageSize := c.QueryPageSize(_defaultPageSize)

	return proxy.GetUsers(filter, orderBy, page, pageSize)
}

// Detail get user
func (o *UserController) Detail(c *web.Ctx) (web.Any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := validator.Uint64("id", id); err != nil {
		return nil, err
	}

	return proxy.GetUser(id)
}

// CreateID create user.ID
func (o *UserController) CreateID(c *web.Ctx) (web.Any, error) {
	return proxy.CreateUserID()
}

// Create create user
func (o *UserController) Create(c *web.Ctx) (web.Any, error) {

	user := model.CreateUser()

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

	var err error

	user := model.CreateUser()

	if err = c.TryParseBody(user); err != nil {
		return nil, err
	}

	if user.ID, err = c.ParamUint64("id"); err != nil {
		return nil, err
	}

	if err = validator.UpdateUser(user); err != nil {
		return nil, err
	}

	return proxy.UpdateUser(user)
}

// Patch update user
func (o *UserController) Patch(c *web.Ctx) (web.Any, error) {

	attrs := c.HeaderAttrs()

	if err := validator.Int(web.HeaderAttrs, len(attrs)); err != nil {
		return nil, err
	}

	var err error

	user := model.CreateUser()

	if err = c.TryParseBody(user); err != nil {
		return nil, err
	}

	if user.ID, err = c.ParamUint64("id"); err != nil {
		return nil, err
	}

	if err = validator.PatchUser(user, attrs...); err != nil {
		return nil, err
	}

	return proxy.PatchUser(user, attrs...)
}

// UpdateStatus update user.Status
func (o *UserController) UpdateStatus(c *web.Ctx) (web.Any, error) {

	var err error

	user := model.CreateUser()

	if err = c.TryParseBody(user); err != nil {
		return nil, err
	}

	if user.ID, err = c.ParamUint64("id"); err != nil {
		return nil, err
	}

	if err = validator.UpdateUserStatus(user); err != nil {
		return nil, err
	}

	return proxy.UpdateUserStatus(user)
}

// Destroy delete user
func (o *UserController) Destroy(c *web.Ctx) (web.Any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := validator.Uint64("id", id); err != nil {
		return nil, err
	}

	return proxy.DestroyUserSoft(id)
}

// Applications return *model.ApplicationCollection, error
func (o *UserController) Applications(c *web.Ctx) (web.Any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	filter := c.QueryFilter()
	orderBy := c.QueryOrderBy()
	page := c.QueryPage(_defaultPage)
	pageSize := c.QueryPageSize(_defaultPageSize)

	return proxy.GetApplicationsByUserID(id, filter, orderBy, page, pageSize)
}

// LinkApplications return rowsAffected int64, error
func (o *UserController) LinkApplications(c *web.Ctx) (web.Any, error) {

	var (
		applicationID []uint64
	)

	id, err := c.ParamUint64("id")

	if err != nil {
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
		applicationID []uint64
	)

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := c.TryParseBody(&applicationID); err != nil {
		return nil, err
	}

	return proxy.UnLinkUserApplications(id, applicationID...)
}

// Application return *model.ApplicationUser, error
func (o *UserController) Application(c *web.Ctx) (web.Any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	applicationID, err := c.ParamUint64("applicationID")

	if err != nil {
		return nil, err
	}

	return proxy.GetUserApplication(id, applicationID)
}

// UpdateApplication return rowsAffected int64, error
func (o *UserController) UpdateApplication(c *web.Ctx) (web.Any, error) {

	applicationUser := model.CreateApplicationUser()

	if err := c.TryParseBody(applicationUser); err != nil {
		return nil, err
	}

	var err error

	applicationUser.ApplicationID, err = c.ParamUint64("applicationID")

	if err != nil {
		return nil, err
	}

	applicationUser.UserID, err = c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := validator.UpdateUserApplication(applicationUser); err != nil {
		return nil, err
	}

	return proxy.UpdateUserApplication(applicationUser)
}

// Roles return *model.RoleCollection, error
func (o *UserController) Roles(c *web.Ctx) (web.Any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	filter := c.QueryFilter()
	orderBy := c.QueryOrderBy()
	page := c.QueryPage(_defaultPage)
	pageSize := c.QueryPageSize(_defaultPageSize)

	return proxy.GetRolesByUserID(id, filter, orderBy, page, pageSize)
}

// LinkRoles return rowsAffected int64, error
func (o *UserController) LinkRoles(c *web.Ctx) (web.Any, error) {

	var (
		roleID []uint64
	)

	id, err := c.ParamUint64("id")

	if err != nil {
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
		roleID []uint64
	)

	id, err := c.ParamUint64("id")

	if err != nil {
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