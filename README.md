# Web.go The library for web

### Route
```go
package route

import (
	"app.gostartkit.com/go/auth/config"
	"app.gostartkit.com/go/auth/controller"
	"app.gostartkit.com/go/auth/middleware"
	"pkg.gostartkit.com/web"
)

func userRoute(app *web.Application, prefix string) {

	c := controller.CreateUserController()

	app.Get(prefix+"/user/", middleware.Chain(c.Index, config.Read|config.ReadUser))
	app.Get(prefix+"/user/:id", middleware.Chain(c.Detail, config.Read|config.ReadUser))
	app.Post(prefix+"/apply/user/id/", middleware.Chain(c.CreateID, config.Write|config.WriteUser))
	app.Post(prefix+"/user/", middleware.Chain(c.Create, config.Write|config.WriteUser))
	app.Put(prefix+"/user/:id", middleware.Chain(c.Update, config.Write|config.WriteUser))
	app.Patch(prefix+"/user/:id", middleware.Chain(c.Patch, config.Write|config.WriteUser))
	app.Patch(prefix+"/user/:id/status/", middleware.Chain(c.UpdateStatus, config.Write|config.WriteUser))
	app.Delete(prefix+"/user/:id", middleware.Chain(c.Destroy, config.Write|config.WriteUser))
	app.Get(prefix+"/user/:id/application/", middleware.Chain(c.Applications, config.Read|config.ReadUser))
	app.Post(prefix+"/user/:id/application/", middleware.Chain(c.LinkApplications, config.Write|config.WriteUser))
	app.Delete(prefix+"/user/:id/application/", middleware.Chain(c.UnLinkApplications, config.Write|config.WriteUser))
	app.Get(prefix+"/user/:id/application/:applicationID", middleware.Chain(c.Application, config.Read|config.ReadUser))
	app.Put(prefix+"/user/:id/application/:applicationID", middleware.Chain(c.UpdateApplication, config.Write|config.WriteUser))
	app.Get(prefix+"/user/:id/role/", middleware.Chain(c.Roles, config.Read|config.ReadUser))
	app.Post(prefix+"/user/:id/role/", middleware.Chain(c.LinkRoles, config.Write|config.WriteUser))
	app.Delete(prefix+"/user/:id/role/", middleware.Chain(c.UnLinkRoles, config.Write|config.WriteUser))
}

```

### Controller
```go
package controller

import (
	"sync"

	"app.gostartkit.com/go/auth/model"
	"app.gostartkit.com/go/auth/proxy"
	"app.gostartkit.com/go/auth/validator"
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
func (r *UserController) Index(c *web.Ctx) (any, error) {

	filter := c.QueryFilter()
	orderBy := c.QueryOrderBy()
	page := c.QueryPage(_defaultPage)
	pageSize := c.QueryPageSize(_defaultPageSize)

	return proxy.GetUsers(filter, orderBy, page, pageSize)
}

// Detail get user
func (r *UserController) Detail(c *web.Ctx) (any, error) {

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
func (r *UserController) CreateID(c *web.Ctx) (any, error) {
	return proxy.CreateUserId()
}

// Create create user
func (r *UserController) Create(c *web.Ctx) (any, error) {

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
func (r *UserController) Update(c *web.Ctx) (any, error) {

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
func (r *UserController) Patch(c *web.Ctx) (any, error) {

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
func (r *UserController) UpdateStatus(c *web.Ctx) (any, error) {

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
func (r *UserController) Destroy(c *web.Ctx) (any, error) {

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
func (r *UserController) Applications(c *web.Ctx) (any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	filter := c.QueryFilter()
	orderBy := c.QueryOrderBy()
	page := c.QueryPage(_defaultPage)
	pageSize := c.QueryPageSize(_defaultPageSize)

	return proxy.GetApplicationsByUserId(id, filter, orderBy, page, pageSize)
}

// LinkApplications return rowsAffected int64, error
func (r *UserController) LinkApplications(c *web.Ctx) (any, error) {

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
func (r *UserController) UnLinkApplications(c *web.Ctx) (any, error) {

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
func (r *UserController) Application(c *web.Ctx) (any, error) {

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
func (r *UserController) UpdateApplication(c *web.Ctx) (any, error) {

	applicationUser := model.CreateApplicationUser()

	if err := c.TryParseBody(applicationUser); err != nil {
		return nil, err
	}

	var err error

	applicationUser.ApplicationID, err = c.ParamUint64("applicationID")

	if err != nil {
		return nil, err
	}

	applicationUser.UserId, err = c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	if err := validator.UpdateUserApplication(applicationUser); err != nil {
		return nil, err
	}

	return proxy.UpdateUserApplication(applicationUser)
}

// Roles return *model.RoleCollection, error
func (r *UserController) Roles(c *web.Ctx) (any, error) {

	id, err := c.ParamUint64("id")

	if err != nil {
		return nil, err
	}

	filter := c.QueryFilter()
	orderBy := c.QueryOrderBy()
	page := c.QueryPage(_defaultPage)
	pageSize := c.QueryPageSize(_defaultPageSize)

	return proxy.GetRolesByUserId(id, filter, orderBy, page, pageSize)
}

// LinkRoles return rowsAffected int64, error
func (r *UserController) LinkRoles(c *web.Ctx) (any, error) {

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
func (r *UserController) UnLinkRoles(c *web.Ctx) (any, error) {

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

### Acknowledgments

Thanks to all open-source projects, I’ve learned a lot from them.

Special thanks to：

- [httprouter](https://github.com/julienschmidt/httprouter): A high-performance HTTP router that inspired the routing logic in this project.
- [web](https://github.com/hoisie/web): A lightweight web framework that provided insights into efficient server design.