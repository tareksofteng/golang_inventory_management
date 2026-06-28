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

type SaleItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64 `json:"unit_price" binding:"gte=0"`
}

type CreateSaleRequest struct {
	CustomerID uint              `json:"customer_id" binding:"required"`
	Discount   float64           `json:"discount" binding:"gte=0"`
	TaxPercent float64           `json:"tax_percent" binding:"gte=0"`
	PaidAmount float64           `json:"paid_amount" binding:"gte=0"`
	Note       string            `json:"note" binding:"max=255"`
	Items      []SaleItemRequest `json:"items" binding:"required,min=1,dive"`
}

type SaleController struct {
	service services.SaleService
}

func NewSaleController(service services.SaleService) *SaleController {
	return &SaleController{service: service}
}

// Create godoc
// @Summary  Create a sale invoice (decreases stock + customer due, transactional, oversell-guarded)
// @Tags     Sales
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreateSaleRequest  true  "Sale with line items"
// @Success  201   {object}  map[string]interface{}
// @Failure  422   {object}  map[string]interface{}  "insufficient stock"
// @Router   /sales [post]
func (ctrl *SaleController) Create(c *gin.Context) {
	var req CreateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	items := make([]services.SaleItemInput, len(req.Items))
	for i, it := range req.Items {
		items[i] = services.SaleItemInput{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
		}
	}

	sale, err := ctrl.service.Create(services.CreateSaleInput{
		CustomerID: req.CustomerID,
		UserID:     middleware.UserID(c),
		Discount:   req.Discount,
		TaxPercent: req.TaxPercent,
		PaidAmount: req.PaidAmount,
		Note:       req.Note,
		Items:      items,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCustomerNotFound):
			response.Error(c, http.StatusUnprocessableEntity, "customer_id does not exist", nil)
		case errors.Is(err, services.ErrProductNotFound):
			response.Error(c, http.StatusUnprocessableEntity, "one of the products does not exist", nil)
		case errors.Is(err, services.ErrInsufficientStock):
			response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
		case errors.Is(err, services.ErrPaidExceedsTotal):
			response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
		case errors.Is(err, services.ErrNoItems):
			response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			response.InternalError(c, "Failed to create sale")
		}
		return
	}
	response.Created(c, "Sale created successfully", sale)
}

// List godoc
// @Summary  List sales (paginated)
// @Tags     Sales
// @Produce  json
// @Security BearerAuth
// @Param    page      query     int     false  "Page number"
// @Param    per_page  query     int     false  "Items per page"
// @Param    search    query     string  false  "Search by invoice no"
// @Success  200       {object}  map[string]interface{}
// @Router   /sales [get]
func (ctrl *SaleController) List(c *gin.Context) {
	p := pagination.Parse(c)

	sales, total, err := ctrl.service.List(p.Search, p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch sales")
		return
	}

	meta := response.Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: pagination.TotalPages(total, p.PerPage),
	}
	response.Paginated(c, "Sales fetched successfully", sales, meta)
}

func (ctrl *SaleController) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid sale id", nil)
		return
	}

	sale, err := ctrl.service.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrSaleNotFound) {
			response.NotFound(c, "Sale not found")
			return
		}
		response.InternalError(c, "Failed to fetch sale")
		return
	}
	response.Success(c, "Sale fetched successfully", sale)
}
