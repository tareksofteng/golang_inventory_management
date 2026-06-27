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

type ReturnItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	UnitValue float64 `json:"unit_value" binding:"gte=0"` // cost (purchase) or price (sale) per unit
}

type CreatePurchaseReturnRequest struct {
	SupplierID uint                `json:"supplier_id" binding:"required"`
	Note       string              `json:"note" binding:"max=255"`
	Items      []ReturnItemRequest `json:"items" binding:"required,min=1,dive"`
}

type CreateSaleReturnRequest struct {
	CustomerID uint                `json:"customer_id" binding:"required"`
	Note       string              `json:"note" binding:"max=255"`
	Items      []ReturnItemRequest `json:"items" binding:"required,min=1,dive"`
}

type ReturnController struct {
	service services.ReturnService
}

func NewReturnController(service services.ReturnService) *ReturnController {
	return &ReturnController{service: service}
}

func toReturnItems(in []ReturnItemRequest) []services.ReturnItemInput {
	out := make([]services.ReturnItemInput, len(in))
	for i, it := range in {
		out[i] = services.ReturnItemInput{ProductID: it.ProductID, Quantity: it.Quantity, UnitValue: it.UnitValue}
	}
	return out
}

// CreatePurchaseReturn godoc
// @Summary  Return goods to a supplier (decreases stock + supplier due, transactional)
// @Tags     Returns
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreatePurchaseReturnRequest  true  "Purchase return"
// @Success  201   {object}  map[string]interface{}
// @Failure  422   {object}  map[string]interface{}
// @Router   /returns/purchase [post]
func (ctrl *ReturnController) CreatePurchaseReturn(c *gin.Context) {
	var req CreatePurchaseReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	ret, err := ctrl.service.CreatePurchaseReturn(services.CreatePurchaseReturnInput{
		SupplierID: req.SupplierID,
		UserID:     middleware.UserID(c),
		Note:       req.Note,
		Items:      toReturnItems(req.Items),
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSupplierNotFound):
			response.Error(c, http.StatusUnprocessableEntity, "supplier_id does not exist", nil)
		case errors.Is(err, services.ErrProductNotFound):
			response.Error(c, http.StatusUnprocessableEntity, "one of the products does not exist", nil)
		case errors.Is(err, services.ErrInsufficientStock):
			response.Error(c, http.StatusUnprocessableEntity, "cannot return more than current stock", nil)
		case errors.Is(err, services.ErrNoItems):
			response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			response.InternalError(c, "Failed to create purchase return")
		}
		return
	}
	response.Created(c, "Purchase return created", ret)
}

// ListPurchaseReturns godoc
// @Summary  List purchase returns (paginated)
// @Tags     Returns
// @Produce  json
// @Security BearerAuth
// @Param    page      query     int  false  "Page number"
// @Param    per_page  query     int  false  "Items per page"
// @Success  200       {object}  map[string]interface{}
// @Router   /returns/purchase [get]
func (ctrl *ReturnController) ListPurchaseReturns(c *gin.Context) {
	p := pagination.Parse(c)
	rows, total, err := ctrl.service.ListPurchaseReturns(p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch purchase returns")
		return
	}
	response.Paginated(c, "Purchase returns", rows, response.Meta{
		Page: p.Page, PerPage: p.PerPage, Total: total, TotalPages: pagination.TotalPages(total, p.PerPage),
	})
}

// CreateSaleReturn godoc
// @Summary  Accept a customer return (increases stock + reduces customer due, transactional)
// @Tags     Returns
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreateSaleReturnRequest  true  "Sale return"
// @Success  201   {object}  map[string]interface{}
// @Failure  422   {object}  map[string]interface{}
// @Router   /returns/sale [post]
func (ctrl *ReturnController) CreateSaleReturn(c *gin.Context) {
	var req CreateSaleReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	ret, err := ctrl.service.CreateSaleReturn(services.CreateSaleReturnInput{
		CustomerID: req.CustomerID,
		UserID:     middleware.UserID(c),
		Note:       req.Note,
		Items:      toReturnItems(req.Items),
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCustomerNotFound):
			response.Error(c, http.StatusUnprocessableEntity, "customer_id does not exist", nil)
		case errors.Is(err, services.ErrProductNotFound):
			response.Error(c, http.StatusUnprocessableEntity, "one of the products does not exist", nil)
		case errors.Is(err, services.ErrNoItems):
			response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			response.InternalError(c, "Failed to create sale return")
		}
		return
	}
	response.Created(c, "Sale return created", ret)
}

// ListSaleReturns godoc
// @Summary  List sale returns (paginated)
// @Tags     Returns
// @Produce  json
// @Security BearerAuth
// @Param    page      query     int  false  "Page number"
// @Param    per_page  query     int  false  "Items per page"
// @Success  200       {object}  map[string]interface{}
// @Router   /returns/sale [get]
func (ctrl *ReturnController) ListSaleReturns(c *gin.Context) {
	p := pagination.Parse(c)
	rows, total, err := ctrl.service.ListSaleReturns(p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch sale returns")
		return
	}
	response.Paginated(c, "Sale returns", rows, response.Meta{
		Page: p.Page, PerPage: p.PerPage, Total: total, TotalPages: pagination.TotalPages(total, p.PerPage),
	})
}
