package controllers

import (
	"errors"

	"inventory-api/internal/services"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type AccountingController struct {
	service services.AccountingService
}

func NewAccountingController(service services.AccountingService) *AccountingController {
	return &AccountingController{service: service}
}

// TrialBalance godoc
// @Summary  Trial balance — every account's net debit/credit with matching totals
// @Tags     Accounting
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  map[string]interface{}
// @Router   /accounting/trial-balance [get]
func (ctrl *AccountingController) TrialBalance(c *gin.Context) {
	tb, err := ctrl.service.TrialBalance()
	if err != nil {
		response.InternalError(c, "Failed to build trial balance")
		return
	}
	response.Success(c, "Trial balance", tb)
}

// GeneralLedger godoc
// @Summary  General ledger for one account (postings + running balance)
// @Tags     Accounting
// @Produce  json
// @Security BearerAuth
// @Param    id   path      int  true  "Account ID"
// @Success  200  {object}  map[string]interface{}
// @Router   /accounting/ledger/{id} [get]
func (ctrl *AccountingController) GeneralLedger(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid account id", nil)
		return
	}
	gl, err := ctrl.service.GeneralLedger(id)
	if err != nil {
		if errors.Is(err, services.ErrAccountNotFound) {
			response.NotFound(c, "Account not found")
			return
		}
		response.InternalError(c, "Failed to build general ledger")
		return
	}
	response.Success(c, "General ledger", gl)
}
