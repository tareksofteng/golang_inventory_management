package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"inventory-api/internal/models"
	"inventory-api/internal/services"
	"inventory-api/pkg/pagination"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ---- Request DTOs ----------------------------------------------------------
//
// These structs define the SHAPE + VALIDATION RULES of incoming JSON. This is
// the Go equivalent of a Laravel Form Request. We keep them separate from the
// model so the API contract and the DB schema can evolve independently.
//
// `binding` tags are validated by Gin automatically on ShouldBindJSON.
// IsActive is a *bool so we can tell "field omitted" (nil) from "false".

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=255"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=255"`
	IsActive    *bool  `json:"is_active"`
}

// ---- Controller ------------------------------------------------------------

// CategoryController is the HTTP layer: it parses requests, calls the service,
// and writes JSON responses. It knows nothing about the DB.
type CategoryController struct {
	service services.CategoryService
}

// NewCategoryController injects the service (an interface) into the controller.
func NewCategoryController(service services.CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

// Create godoc
// @Summary  Create a category
// @Tags     Categories
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreateCategoryRequest  true  "Category"
// @Success  201   {object}  map[string]interface{}
// @Router   /categories [post]
func (ctrl *CategoryController) Create(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	category := &models.Category{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true, // default
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	created, err := ctrl.service.Create(category)
	if err != nil {
		if errors.Is(err, services.ErrCategoryNameTaken) {
			response.Error(c, http.StatusConflict, err.Error(), nil)
			return
		}
		response.InternalError(c, "Failed to create category")
		return
	}

	response.Created(c, "Category created successfully", created)
}

// List godoc
// @Summary  List categories (paginated, searchable)
// @Tags     Categories
// @Produce  json
// @Security BearerAuth
// @Param    page      query     int     false  "Page number"
// @Param    per_page  query     int     false  "Items per page"
// @Param    search    query     string  false  "Search by name"
// @Success  200       {object}  map[string]interface{}
// @Router   /categories [get]
func (ctrl *CategoryController) List(c *gin.Context) {
	p := pagination.Parse(c)

	categories, total, err := ctrl.service.List(p.Search, p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch categories")
		return
	}

	meta := response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, p.PerPage),
	}
	response.Paginated(c, "Categories fetched successfully", categories, meta)
}

// Get handles GET /categories/:id
func (ctrl *CategoryController) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid category id", nil)
		return
	}

	category, err := ctrl.service.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrCategoryNotFound) {
			response.NotFound(c, "Category not found")
			return
		}
		response.InternalError(c, "Failed to fetch category")
		return
	}

	response.Success(c, "Category fetched successfully", category)
}

// Update handles PUT /categories/:id
func (ctrl *CategoryController) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid category id", nil)
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	data := &models.Category{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}
	if req.IsActive != nil {
		data.IsActive = *req.IsActive
	}

	updated, err := ctrl.service.Update(id, data)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCategoryNotFound):
			response.NotFound(c, "Category not found")
		case errors.Is(err, services.ErrCategoryNameTaken):
			response.Error(c, http.StatusConflict, err.Error(), nil)
		default:
			response.InternalError(c, "Failed to update category")
		}
		return
	}

	response.Success(c, "Category updated successfully", updated)
}

// Delete handles DELETE /categories/:id (soft delete)
func (ctrl *CategoryController) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid category id", nil)
		return
	}

	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, services.ErrCategoryNotFound) {
			response.NotFound(c, "Category not found")
			return
		}
		response.InternalError(c, "Failed to delete category")
		return
	}

	response.Success(c, "Category deleted successfully", nil)
}

// ---- Shared helper (package-level, reused by every controller) -------------

// parseIDParam reads the :id path segment and validates it is a positive
// integer. All controllers in this package share this single helper.
func parseIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}
