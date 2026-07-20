package controllers

import (
	"errors"
	"net/http"

	"inventory-api/internal/models"
	"inventory-api/internal/services"
	"inventory-api/pkg/pagination"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type CreateAccountRequest struct {
	Code     string `json:"code" binding:"required,max=20"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Type     string `json:"type" binding:"required,oneof=asset liability equity income expense"`
	IsActive *bool  `json:"is_active"`
}

type UpdateAccountRequest struct {
	Code     string `json:"code" binding:"required,max=20"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Type     string `json:"type" binding:"required,oneof=asset liability equity income expense"`
	IsActive *bool  `json:"is_active"`
}

type AccountController struct {
	service services.AccountService
}

func NewAccountController(service services.AccountService) *AccountController {
	return &AccountController{service: service}
}

func handleAccountWriteError(c *gin.Context, err error, action string) {
	switch {
	case errors.Is(err, services.ErrAccountNotFound):
		response.NotFound(c, "Account not found")
	case errors.Is(err, services.ErrAccountCodeTaken):
		response.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, services.ErrInvalidAccountType):
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
	default:
		response.InternalError(c, "Failed to "+action+" account")
	}
}

// Create godoc
// @Summary  Create a chart-of-accounts account
// @Tags     Accounts
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreateAccountRequest  true  "Account"
// @Success  201   {object}  map[string]interface{}
// @Router   /accounts [post]
func (ctrl *AccountController) Create(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}
	account := &models.Account{Code: req.Code, Name: req.Name, Type: req.Type, IsActive: true}
	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}
	created, err := ctrl.service.Create(account)
	if err != nil {
		handleAccountWriteError(c, err, "create")
		return
	}
	response.Created(c, "Account created successfully", created)
}

// List godoc
// @Summary  List accounts (paginated, filterable by type)
// @Tags     Accounts
// @Produce  json
// @Security BearerAuth
// @Param    page      query     int     false  "Page number"
// @Param    per_page  query     int     false  "Items per page"
// @Param    search    query     string  false  "Search by code or name"
// @Param    type      query     string  false  "Filter by type"
// @Success  200       {object}  map[string]interface{}
// @Router   /accounts [get]
func (ctrl *AccountController) List(c *gin.Context) {
	p := pagination.Parse(c)
	accType := c.Query("type")

	accounts, total, err := ctrl.service.List(p.Search, accType, p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch accounts")
		return
	}
	response.Paginated(c, "Accounts fetched successfully", accounts, response.Meta{
		Page: p.Page, PerPage: p.PerPage, Total: total, TotalPages: pagination.TotalPages(total, p.PerPage),
	})
}

// Active godoc
// @Summary  List all active accounts (for journal dropdowns)
// @Tags     Accounts
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  map[string]interface{}
// @Router   /accounts/active [get]
func (ctrl *AccountController) Active(c *gin.Context) {
	accounts, err := ctrl.service.ListActive()
	if err != nil {
		response.InternalError(c, "Failed to fetch accounts")
		return
	}
	response.Success(c, "Active accounts", accounts)
}

// Get godoc
// @Summary  Get an account by id
// @Tags     Accounts
// @Produce  json
// @Security BearerAuth
// @Param    id   path      int  true  "Account ID"
// @Success  200  {object}  map[string]interface{}
// @Router   /accounts/{id} [get]
func (ctrl *AccountController) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid account id", nil)
		return
	}
	account, err := ctrl.service.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrAccountNotFound) {
			response.NotFound(c, "Account not found")
			return
		}
		response.InternalError(c, "Failed to fetch account")
		return
	}
	response.Success(c, "Account fetched successfully", account)
}

// Update godoc
// @Summary  Update an account
// @Tags     Accounts
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id    path      int                  true  "Account ID"
// @Param    body  body      UpdateAccountRequest true  "Account"
// @Success  200   {object}  map[string]interface{}
// @Router   /accounts/{id} [put]
func (ctrl *AccountController) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid account id", nil)
		return
	}

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	account, err := ctrl.service.Update(id, &models.Account{
		Code:     req.Code,
		Name:     req.Name,
		Type:     req.Type,
		IsActive: isActive,
	})
	if err != nil {
		handleAccountWriteError(c, err, "update")
		return
	}

	response.Success(c, "Account updated successfully", account)
}

// Delete godoc
// @Summary  Delete an account (soft delete)
// @Tags     Accounts
// @Produce  json
// @Security BearerAuth
// @Param    id   path      int  true  "Account ID"
// @Success  200  {object}  map[string]interface{}
// @Router   /accounts/{id} [delete]
func (ctrl *AccountController) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid account id", nil)
		return
	}
	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, services.ErrAccountNotFound) {
			response.NotFound(c, "Account not found")
			return
		}
		response.InternalError(c, "Failed to delete account")
		return
	}
	response.Success(c, "Account deleted successfully", nil)
}
