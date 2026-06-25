package controllers

import (
	"errors"

	"inventory-api/internal/models"
	"inventory-api/internal/services"
	"inventory-api/pkg/pagination"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type CreateCustomerRequest struct {
	Name     string  `json:"name" binding:"required,min=2,max=100"`
	Email    string  `json:"email" binding:"omitempty,email,max=100"`
	Phone    string  `json:"phone" binding:"max=20"`
	Address  string  `json:"address" binding:"max=255"`
	Due      float64 `json:"due" binding:"gte=0"`
	IsActive *bool   `json:"is_active"`
}

type UpdateCustomerRequest struct {
	Name     string  `json:"name" binding:"required,min=2,max=100"`
	Email    string  `json:"email" binding:"omitempty,email,max=100"`
	Phone    string  `json:"phone" binding:"max=20"`
	Address  string  `json:"address" binding:"max=255"`
	Due      float64 `json:"due" binding:"gte=0"`
	IsActive *bool   `json:"is_active"`
}

type CustomerController struct {
	service services.CustomerService
}

func NewCustomerController(service services.CustomerService) *CustomerController {
	return &CustomerController{service: service}
}

func (ctrl *CustomerController) Create(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	customer := &models.Customer{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Address:  req.Address,
		Due:      req.Due,
		IsActive: true,
	}
	if req.IsActive != nil {
		customer.IsActive = *req.IsActive
	}

	created, err := ctrl.service.Create(customer)
	if err != nil {
		response.InternalError(c, "Failed to create customer")
		return
	}
	response.Created(c, "Customer created successfully", created)
}

func (ctrl *CustomerController) List(c *gin.Context) {
	p := pagination.Parse(c)

	customers, total, err := ctrl.service.List(p.Search, p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch customers")
		return
	}

	meta := response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, p.PerPage),
	}
	response.Paginated(c, "Customers fetched successfully", customers, meta)
}

func (ctrl *CustomerController) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid customer id", nil)
		return
	}

	customer, err := ctrl.service.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrCustomerNotFound) {
			response.NotFound(c, "Customer not found")
			return
		}
		response.InternalError(c, "Failed to fetch customer")
		return
	}
	response.Success(c, "Customer fetched successfully", customer)
}

func (ctrl *CustomerController) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid customer id", nil)
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	data := &models.Customer{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Address:  req.Address,
		Due:      req.Due,
		IsActive: true,
	}
	if req.IsActive != nil {
		data.IsActive = *req.IsActive
	}

	updated, err := ctrl.service.Update(id, data)
	if err != nil {
		if errors.Is(err, services.ErrCustomerNotFound) {
			response.NotFound(c, "Customer not found")
			return
		}
		response.InternalError(c, "Failed to update customer")
		return
	}
	response.Success(c, "Customer updated successfully", updated)
}

func (ctrl *CustomerController) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid customer id", nil)
		return
	}

	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, services.ErrCustomerNotFound) {
			response.NotFound(c, "Customer not found")
			return
		}
		response.InternalError(c, "Failed to delete customer")
		return
	}
	response.Success(c, "Customer deleted successfully", nil)
}
