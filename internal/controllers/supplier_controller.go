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

// ---- Request DTOs ----------------------------------------------------------
// Email is required + must be a valid email: it is the unique field, and
// requiring it avoids the "multiple empty emails collide on the unique index"
// problem we discussed in the repository.

type CreateSupplierRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Phone    string `json:"phone" binding:"max=20"`
	Address  string `json:"address" binding:"max=255"`
	IsActive *bool  `json:"is_active"`
}

type UpdateSupplierRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Phone    string `json:"phone" binding:"max=20"`
	Address  string `json:"address" binding:"max=255"`
	IsActive *bool  `json:"is_active"`
}

// ---- Controller ------------------------------------------------------------

type SupplierController struct {
	service services.SupplierService
}

func NewSupplierController(service services.SupplierService) *SupplierController {
	return &SupplierController{service: service}
}

// Create handles POST /suppliers
func (ctrl *SupplierController) Create(c *gin.Context) {
	var req CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	supplier := &models.Supplier{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Address:  req.Address,
		IsActive: true,
	}
	if req.IsActive != nil {
		supplier.IsActive = *req.IsActive
	}

	created, err := ctrl.service.Create(supplier)
	if err != nil {
		if errors.Is(err, services.ErrSupplierEmailTaken) {
			response.Error(c, http.StatusConflict, err.Error(), nil)
			return
		}
		response.InternalError(c, "Failed to create supplier")
		return
	}

	response.Created(c, "Supplier created successfully", created)
}

// List handles GET /suppliers?page=&per_page=&search=
func (ctrl *SupplierController) List(c *gin.Context) {
	p := pagination.Parse(c)

	suppliers, total, err := ctrl.service.List(p.Search, p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch suppliers")
		return
	}

	meta := response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, p.PerPage),
	}
	response.Paginated(c, "Suppliers fetched successfully", suppliers, meta)
}

// Get handles GET /suppliers/:id
func (ctrl *SupplierController) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid supplier id", nil)
		return
	}

	supplier, err := ctrl.service.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrSupplierNotFound) {
			response.NotFound(c, "Supplier not found")
			return
		}
		response.InternalError(c, "Failed to fetch supplier")
		return
	}

	response.Success(c, "Supplier fetched successfully", supplier)
}

// Update handles PUT /suppliers/:id
func (ctrl *SupplierController) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid supplier id", nil)
		return
	}

	var req UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	data := &models.Supplier{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Address:  req.Address,
		IsActive: true,
	}
	if req.IsActive != nil {
		data.IsActive = *req.IsActive
	}

	updated, err := ctrl.service.Update(id, data)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSupplierNotFound):
			response.NotFound(c, "Supplier not found")
		case errors.Is(err, services.ErrSupplierEmailTaken):
			response.Error(c, http.StatusConflict, err.Error(), nil)
		default:
			response.InternalError(c, "Failed to update supplier")
		}
		return
	}

	response.Success(c, "Supplier updated successfully", updated)
}

// Delete handles DELETE /suppliers/:id (soft delete)
func (ctrl *SupplierController) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid supplier id", nil)
		return
	}

	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, services.ErrSupplierNotFound) {
			response.NotFound(c, "Supplier not found")
			return
		}
		response.InternalError(c, "Failed to delete supplier")
		return
	}

	response.Success(c, "Supplier deleted successfully", nil)
}
