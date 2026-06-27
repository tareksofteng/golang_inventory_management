package controllers

import (
	"errors"
	"net/http"

	"inventory-api/internal/middleware"
	"inventory-api/internal/services"
	"inventory-api/pkg/pagination"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type CustomerPaymentRequest struct {
	CustomerID uint    `json:"customer_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Method     string  `json:"method" binding:"required,oneof=cash bank mobile"`
	Note       string  `json:"note" binding:"max=255"`
}

type SupplierPaymentRequest struct {
	SupplierID uint    `json:"supplier_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Method     string  `json:"method" binding:"required,oneof=cash bank mobile"`
	Note       string  `json:"note" binding:"max=255"`
}

type PaymentController struct {
	service services.PaymentService
}

func NewPaymentController(service services.PaymentService) *PaymentController {
	return &PaymentController{service: service}
}

func (ctrl *PaymentController) PayCustomer(c *gin.Context) {
	var req CustomerPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	payment, err := ctrl.service.PayCustomer(req.CustomerID, req.Amount, req.Method, req.Note, middleware.UserID(c))
	if err != nil {
		ctrl.handleWriteError(c, err, services.ErrCustomerNotFound, "customer")
		return
	}
	response.Created(c, "Customer payment received", payment)
}

func (ctrl *PaymentController) PaySupplier(c *gin.Context) {
	var req SupplierPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	payment, err := ctrl.service.PaySupplier(req.SupplierID, req.Amount, req.Method, req.Note, middleware.UserID(c))
	if err != nil {
		ctrl.handleWriteError(c, err, services.ErrSupplierNotFound, "supplier")
		return
	}
	response.Created(c, "Supplier payment recorded", payment)
}

// handleWriteError maps payment errors to HTTP codes. notFound is the
// party-specific sentinel (customer/supplier) for this call.
func (ctrl *PaymentController) handleWriteError(c *gin.Context, err, notFound error, party string) {
	switch {
	case errors.Is(err, notFound):
		response.Error(c, http.StatusUnprocessableEntity, party+" does not exist", nil)
	case errors.Is(err, services.ErrAmountExceedsDue):
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
	default:
		response.InternalError(c, "Failed to record payment")
	}
}

func (ctrl *PaymentController) List(c *gin.Context) {
	p := pagination.Parse(c)
	partyType := c.Query("type") // optional: customer | supplier

	payments, total, err := ctrl.service.List(partyType, p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch payments")
		return
	}

	meta := response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, p.PerPage),
	}
	response.Paginated(c, "Payments fetched successfully", payments, meta)
}
