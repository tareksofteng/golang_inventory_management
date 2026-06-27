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
// Quantity is accepted on create (opening stock) but NOT on update — stock
// only moves via the Stock In module.

type CreateProductRequest struct {
	Name       string  `json:"name" binding:"required,min=2,max=150"`
	SKU        string  `json:"sku" binding:"required,max=50"`
	CategoryID uint    `json:"category_id" binding:"required"`
	SupplierID uint    `json:"supplier_id" binding:"required"`
	Price      float64 `json:"price" binding:"gte=0"`
	CostPrice  float64 `json:"cost_price" binding:"gte=0"`
	Quantity   int     `json:"quantity" binding:"gte=0"`
	Unit       string  `json:"unit" binding:"max=20"`
	IsActive   *bool   `json:"is_active"`
}

type UpdateProductRequest struct {
	Name       string  `json:"name" binding:"required,min=2,max=150"`
	SKU        string  `json:"sku" binding:"required,max=50"`
	CategoryID uint    `json:"category_id" binding:"required"`
	SupplierID uint    `json:"supplier_id" binding:"required"`
	Price      float64 `json:"price" binding:"gte=0"`
	CostPrice  float64 `json:"cost_price" binding:"gte=0"`
	Unit       string  `json:"unit" binding:"max=20"`
	IsActive   *bool   `json:"is_active"`
}

// ---- Controller ------------------------------------------------------------

type ProductController struct {
	service services.ProductService
}

func NewProductController(service services.ProductService) *ProductController {
	return &ProductController{service: service}
}

// handleWriteError maps service errors shared by create/update to HTTP codes.
func handleProductWriteError(c *gin.Context, err error, action string) {
	switch {
	case errors.Is(err, services.ErrProductNotFound):
		response.NotFound(c, "Product not found")
	case errors.Is(err, services.ErrProductSKUTaken):
		response.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, services.ErrCategoryNotFound):
		response.Error(c, http.StatusUnprocessableEntity, "category_id does not exist", nil)
	case errors.Is(err, services.ErrSupplierNotFound):
		response.Error(c, http.StatusUnprocessableEntity, "supplier_id does not exist", nil)
	default:
		response.InternalError(c, "Failed to "+action+" product")
	}
}

// Create godoc
// @Summary  Create a product
// @Tags     Products
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreateProductRequest  true  "Product"
// @Success  201   {object}  map[string]interface{}
// @Router   /products [post]
func (ctrl *ProductController) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	product := &models.Product{
		Name:       req.Name,
		SKU:        req.SKU,
		CategoryID: req.CategoryID,
		SupplierID: req.SupplierID,
		Price:      req.Price,
		CostPrice:  req.CostPrice,
		Quantity:   req.Quantity,
		Unit:       req.Unit,
		IsActive:   true,
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	created, err := ctrl.service.Create(product)
	if err != nil {
		handleProductWriteError(c, err, "create")
		return
	}
	response.Created(c, "Product created successfully", created)
}

// List godoc
// @Summary  List products (paginated, searchable)
// @Tags     Products
// @Produce  json
// @Security BearerAuth
// @Param    page      query     int     false  "Page number"
// @Param    per_page  query     int     false  "Items per page"
// @Param    search    query     string  false  "Search by name or SKU"
// @Success  200       {object}  map[string]interface{}
// @Router   /products [get]
func (ctrl *ProductController) List(c *gin.Context) {
	p := pagination.Parse(c)

	products, total, err := ctrl.service.List(p.Search, p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch products")
		return
	}

	meta := response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, p.PerPage),
	}
	response.Paginated(c, "Products fetched successfully", products, meta)
}

// Get godoc
// @Summary  Get a product by id
// @Tags     Products
// @Produce  json
// @Security BearerAuth
// @Param    id   path      int  true  "Product ID"
// @Success  200  {object}  map[string]interface{}
// @Failure  404  {object}  map[string]interface{}
// @Router   /products/{id} [get]
func (ctrl *ProductController) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid product id", nil)
		return
	}

	product, err := ctrl.service.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			response.NotFound(c, "Product not found")
			return
		}
		response.InternalError(c, "Failed to fetch product")
		return
	}
	response.Success(c, "Product fetched successfully", product)
}

// Update godoc
// @Summary  Update a product
// @Tags     Products
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id    path      int                   true  "Product ID"
// @Param    body  body      UpdateProductRequest  true  "Product"
// @Success  200   {object}  map[string]interface{}
// @Router   /products/{id} [put]
func (ctrl *ProductController) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid product id", nil)
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	data := &models.Product{
		Name:       req.Name,
		SKU:        req.SKU,
		CategoryID: req.CategoryID,
		SupplierID: req.SupplierID,
		Price:      req.Price,
		CostPrice:  req.CostPrice,
		Unit:       req.Unit,
		IsActive:   true,
	}
	if req.IsActive != nil {
		data.IsActive = *req.IsActive
	}

	updated, err := ctrl.service.Update(id, data)
	if err != nil {
		handleProductWriteError(c, err, "update")
		return
	}
	response.Success(c, "Product updated successfully", updated)
}

// Delete godoc
// @Summary  Delete a product (soft delete)
// @Tags     Products
// @Produce  json
// @Security BearerAuth
// @Param    id   path      int  true  "Product ID"
// @Success  200  {object}  map[string]interface{}
// @Router   /products/{id} [delete]
func (ctrl *ProductController) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid product id", nil)
		return
	}

	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			response.NotFound(c, "Product not found")
			return
		}
		response.InternalError(c, "Failed to delete product")
		return
	}
	response.Success(c, "Product deleted successfully", nil)
}
