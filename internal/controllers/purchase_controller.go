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

type PurchaseItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	UnitCost  float64 `json:"unit_cost" binding:"gte=0"`
}

type CreatePurchaseRequest struct {
	SupplierID uint                  `json:"supplier_id" binding:"required"`
	Discount   float64               `json:"discount" binding:"gte=0"`
	TaxPercent float64               `json:"tax_percent" binding:"gte=0"`
	PaidAmount float64               `json:"paid_amount" binding:"gte=0"`
	Note       string                `json:"note" binding:"max=255"`
	Items      []PurchaseItemRequest `json:"items" binding:"required,min=1,dive"`
}

type PurchaseController struct {
	service services.PurchaseService
}

func NewPurchaseController(service services.PurchaseService) *PurchaseController {
	return &PurchaseController{service: service}
}

// Create godoc
// @Summary  Create a purchase invoice (increases stock + supplier due, transactional)
// @Tags     Purchases
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreatePurchaseRequest  true  "Purchase with line items"
// @Success  201   {object}  map[string]interface{}
// @Router   /purchases [post]
func (ctrl *PurchaseController) Create(c *gin.Context) {
	var req CreatePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	items := make([]services.PurchaseItemInput, len(req.Items))
	for i, it := range req.Items {
		items[i] = services.PurchaseItemInput{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitCost:  it.UnitCost,
		}
	}

	purchase, err := ctrl.service.Create(services.CreatePurchaseInput{
		SupplierID: req.SupplierID,
		UserID:     middleware.UserID(c), // who is logged in
		Discount:   req.Discount,
		TaxPercent: req.TaxPercent,
		PaidAmount: req.PaidAmount,
		Note:       req.Note,
		Items:      items,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSupplierNotFound):
			response.Error(c, http.StatusUnprocessableEntity, "supplier_id does not exist", nil)
		case errors.Is(err, services.ErrProductNotFound):
			response.Error(c, http.StatusUnprocessableEntity, "one of the products does not exist", nil)
		case errors.Is(err, services.ErrPaidExceedsTotal):
			response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
		case errors.Is(err, services.ErrNoItems):
			response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			response.InternalError(c, "Failed to create purchase")
		}
		return
	}
	response.Created(c, "Purchase created successfully", purchase)
}

// List godoc
// @Summary  List purchases (paginated)
// @Tags     Purchases
// @Produce  json
// @Security BearerAuth
// @Param    page      query     int     false  "Page number"
// @Param    per_page  query     int     false  "Items per page"
// @Param    search    query     string  false  "Search by invoice no"
// @Success  200       {object}  map[string]interface{}
// @Router   /purchases [get]
func (ctrl *PurchaseController) List(c *gin.Context) {
	p := pagination.Parse(c)

	purchases, total, err := ctrl.service.List(p.Search, p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch purchases")
		return
	}

	meta := response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, p.PerPage),
	}
	response.Paginated(c, "Purchases fetched successfully", purchases, meta)
}

func (ctrl *PurchaseController) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid purchase id", nil)
		return
	}

	purchase, err := ctrl.service.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrPurchaseNotFound) {
			response.NotFound(c, "Purchase not found")
			return
		}
		response.InternalError(c, "Failed to fetch purchase")
		return
	}
	response.Success(c, "Purchase fetched successfully", purchase)
}

// Delete godoc
// @Summary  Void a purchase (reverses stock + supplier due, transactional)
// @Tags     Purchases
// @Produce  json
// @Security BearerAuth
// @Param    id   path      int  true  "Purchase ID"
// @Success  200  {object}  map[string]interface{}
// @Router   /purchases/{id} [delete]
func (ctrl *PurchaseController) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid purchase id", nil)
		return
	}
	if err := ctrl.service.Delete(id); err != nil {
		if errors.Is(err, services.ErrPurchaseNotFound) {
			response.NotFound(c, "Purchase not found")
			return
		}
		response.InternalError(c, "Failed to void purchase")
		return
	}
	response.Success(c, "Purchase voided successfully", nil)
}
