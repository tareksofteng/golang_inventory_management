package controllers

import (
	"errors"
	"net/http"

	"inventory-api/internal/services"
	"inventory-api/pkg/pagination"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	Role     string `json:"role" binding:"required,oneof=super_admin admin manager salesman"`
}

type UpdateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Role     string `json:"role" binding:"required,oneof=super_admin admin manager salesman"`
	IsActive *bool  `json:"is_active"`
}

type ChangePasswordRequest struct {
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type UserController struct {
	service services.UserService
}

func NewUserController(service services.UserService) *UserController {
	return &UserController{service: service}
}

func handleUserWriteError(c *gin.Context, err error, action string) {
	switch {
	case errors.Is(err, services.ErrUserNotFound):
		response.NotFound(c, "User not found")
	case errors.Is(err, services.ErrUserEmailTaken):
		response.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, services.ErrInvalidRole):
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
	default:
		response.InternalError(c, "Failed to "+action+" user")
	}
}

// Create godoc
// @Summary  Create a user (requires user.manage)
// @Tags     Users
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreateUserRequest  true  "User"
// @Success  201   {object}  map[string]interface{}
// @Router   /users [post]
func (ctrl *UserController) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	user, err := ctrl.service.Create(req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		handleUserWriteError(c, err, "create")
		return
	}
	response.Created(c, "User created successfully", user)
}

// List godoc
// @Summary  List users (requires user.manage)
// @Tags     Users
// @Produce  json
// @Security BearerAuth
// @Param    page      query     int     false  "Page number"
// @Param    per_page  query     int     false  "Items per page"
// @Param    search    query     string  false  "Search by name or email"
// @Success  200       {object}  map[string]interface{}
// @Router   /users [get]
func (ctrl *UserController) List(c *gin.Context) {
	p := pagination.Parse(c)

	users, total, err := ctrl.service.List(p.Search, p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch users")
		return
	}

	meta := response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, p.PerPage),
	}
	response.Paginated(c, "Users fetched successfully", users, meta)
}

func (ctrl *UserController) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid user id", nil)
		return
	}

	user, err := ctrl.service.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			response.NotFound(c, "User not found")
			return
		}
		response.InternalError(c, "Failed to fetch user")
		return
	}
	response.Success(c, "User fetched successfully", user)
}

func (ctrl *UserController) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid user id", nil)
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user, err := ctrl.service.Update(id, req.Name, req.Email, req.Role, isActive)
	if err != nil {
		handleUserWriteError(c, err, "update")
		return
	}
	response.Success(c, "User updated successfully", user)
}

func (ctrl *UserController) ChangePassword(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid user id", nil)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	if err := ctrl.service.ChangePassword(id, req.Password); err != nil {
		handleUserWriteError(c, err, "update password for")
		return
	}
	response.Success(c, "Password changed successfully", nil)
}

func (ctrl *UserController) Disable(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid user id", nil)
		return
	}

	if err := ctrl.service.Disable(id); err != nil {
		handleUserWriteError(c, err, "disable")
		return
	}
	response.Success(c, "User disabled successfully", nil)
}
