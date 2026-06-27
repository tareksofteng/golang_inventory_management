package controllers

import (
	"errors"

	"inventory-api/internal/services"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type LedgerController struct {
	service services.LedgerService
}

func NewLedgerController(service services.LedgerService) *LedgerController {
	return &LedgerController{service: service}
}

// Customer godoc
// @Summary  Customer ledger — full statement with running balance
// @Tags     Ledger
// @Produce  json
// @Security BearerAuth
// @Param    id   path      int  true  "Customer ID"
// @Success  200  {object}  map[string]interface{}
// @Failure  404  {object}  map[string]interface{}
// @Router   /ledger/customer/{id} [get]
func (ctrl *LedgerController) Customer(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid customer id", nil)
		return
	}

	ledger, err := ctrl.service.CustomerLedger(id)
	if err != nil {
		if errors.Is(err, services.ErrCustomerNotFound) {
			response.NotFound(c, "Customer not found")
			return
		}
		response.InternalError(c, "Failed to build customer ledger")
		return
	}
	response.Success(c, "Customer ledger", ledger)
}

// Supplier godoc
// @Summary  Supplier ledger — full statement with running balance
// @Tags     Ledger
// @Produce  json
// @Security BearerAuth
// @Param    id   path      int  true  "Supplier ID"
// @Success  200  {object}  map[string]interface{}
// @Failure  404  {object}  map[string]interface{}
// @Router   /ledger/supplier/{id} [get]
func (ctrl *LedgerController) Supplier(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid supplier id", nil)
		return
	}

	ledger, err := ctrl.service.SupplierLedger(id)
	if err != nil {
		if errors.Is(err, services.ErrSupplierNotFound) {
			response.NotFound(c, "Supplier not found")
			return
		}
		response.InternalError(c, "Failed to build supplier ledger")
		return
	}
	response.Success(c, "Supplier ledger", ledger)
}
