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
	Quantity  int     `json:"quantity" binding:"gte=0"` // 0 = line not returned
	UnitValue float64 `json:"unit_value" binding:"gte=0"`
}

type CreatePurchaseReturnRequest struct {
	PurchaseID uint                `json:"purchase_id" binding:"required"`
	Note       string              `json:"note" binding:"max=255"`
	Items      []ReturnItemRequest `json:"items" binding:"required,min=1,dive"`
}

type CreateSaleReturnRequest struct {
	SaleID uint                `json:"sale_id" binding:"required"`
	Note   string              `json:"note" binding:"max=255"`
	Items  []ReturnItemRequest `json:"items" binding:"required,min=1,dive"`
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

func returnWriteError(c *gin.Context, err error, notFound error, kind string) {
	switch {
	case errors.Is(err, notFound):
		response.Error(c, http.StatusUnprocessableEntity, kind+" invoice does not exist", nil)
	case errors.Is(err, services.ErrReturnExceedsAvailable):
		response.Error(c, http.StatusUnprocessableEntity, "return quantity exceeds available quantity", nil)
	case errors.Is(err, services.ErrInsufficientStock):
		response.Error(c, http.StatusUnprocessableEntity, "cannot return more than current stock", nil)
	case errors.Is(err, services.ErrNoItems):
		response.Error(c, http.StatusUnprocessableEntity, "enter a return quantity for at least one item", nil)
	default:
		response.InternalError(c, "Failed to create return")
	}
}

// LookupPurchase godoc
// @Summary  Find a purchase by invoice no with returnable quantities
// @Tags     Returns
// @Produce  json
// @Security BearerAuth
// @Param    invoice  query     string  true  "Purchase invoice number"
// @Success  200      {object}  map[string]interface{}
// @Router   /returns/purchase/lookup [get]
func (ctrl *ReturnController) LookupPurchase(c *gin.Context) {
	lookup, err := ctrl.service.LookupPurchase(c.Query("invoice"))
	if err != nil {
		if errors.Is(err, services.ErrPurchaseNotFound) {
			response.NotFound(c, "Purchase invoice not found")
			return
		}
		response.InternalError(c, "Lookup failed")
		return
	}
	response.Success(c, "Purchase found", lookup)
}

// LookupSale godoc
// @Summary  Find a sale by invoice no with returnable quantities
// @Tags     Returns
// @Produce  json
// @Security BearerAuth
// @Param    invoice  query     string  true  "Sale invoice number"
// @Success  200      {object}  map[string]interface{}
// @Router   /returns/sale/lookup [get]
func (ctrl *ReturnController) LookupSale(c *gin.Context) {
	lookup, err := ctrl.service.LookupSale(c.Query("invoice"))
	if err != nil {
		if errors.Is(err, services.ErrSaleNotFound) {
			response.NotFound(c, "Sale invoice not found")
			return
		}
		response.InternalError(c, "Lookup failed")
		return
	}
	response.Success(c, "Sale found", lookup)
}

// CreatePurchaseReturn godoc
// @Summary  Return goods against a purchase invoice (decreases stock + supplier due)
// @Tags     Returns
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreatePurchaseReturnRequest  true  "Purchase return"
// @Success  201   {object}  map[string]interface{}
// @Router   /returns/purchase [post]
func (ctrl *ReturnController) CreatePurchaseReturn(c *gin.Context) {
	var req CreatePurchaseReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}
	ret, err := ctrl.service.CreatePurchaseReturn(services.CreatePurchaseReturnInput{
		PurchaseID: req.PurchaseID,
		UserID:     middleware.UserID(c),
		Note:       req.Note,
		Items:      toReturnItems(req.Items),
	})
	if err != nil {
		returnWriteError(c, err, services.ErrPurchaseNotFound, "purchase")
		return
	}
	response.Created(c, "Purchase return created", ret)
}

// CreateSaleReturn godoc
// @Summary  Return goods against a sale invoice (increases stock + reduces customer due)
// @Tags     Returns
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreateSaleReturnRequest  true  "Sale return"
// @Success  201   {object}  map[string]interface{}
// @Router   /returns/sale [post]
func (ctrl *ReturnController) CreateSaleReturn(c *gin.Context) {
	var req CreateSaleReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}
	ret, err := ctrl.service.CreateSaleReturn(services.CreateSaleReturnInput{
		SaleID: req.SaleID,
		UserID: middleware.UserID(c),
		Note:   req.Note,
		Items:  toReturnItems(req.Items),
	})
	if err != nil {
		returnWriteError(c, err, services.ErrSaleNotFound, "sale")
		return
	}
	response.Created(c, "Sale return created", ret)
}

// ListPurchaseReturns godoc
// @Summary  List purchase returns (paginated)
// @Tags     Returns
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  map[string]interface{}
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

// ListSaleReturns godoc
// @Summary  List sale returns (paginated)
// @Tags     Returns
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  map[string]interface{}
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
