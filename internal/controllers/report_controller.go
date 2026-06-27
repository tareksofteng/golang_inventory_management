package controllers

import (
	"time"

	"inventory-api/internal/services"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type ReportController struct {
	service services.ReportService
}

func NewReportController(service services.ReportService) *ReportController {
	return &ReportController{service: service}
}

// parseDateRange reads ?from=YYYY-MM-DD&to=YYYY-MM-DD. Missing values default
// to the current month. The returned `to` is EXCLUSIVE (next day 00:00) so the
// whole "to" day is included in a `< to` query.
func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	now := time.Now()
	loc := now.Location()

	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	toExcl := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)

	if f := c.Query("from"); f != "" {
		if t, err := time.ParseInLocation("2006-01-02", f, loc); err == nil {
			from = t
		}
	}
	if t2 := c.Query("to"); t2 != "" {
		if t, err := time.ParseInLocation("2006-01-02", t2, loc); err == nil {
			toExcl = t.AddDate(0, 0, 1)
		}
	}
	return from, toExcl
}

func (ctrl *ReportController) Sales(c *gin.Context) {
	from, to := parseDateRange(c)
	report, err := ctrl.service.Sales(from, to)
	if err != nil {
		response.InternalError(c, "Failed to build sales report")
		return
	}
	response.Success(c, "Sales report", report)
}

func (ctrl *ReportController) Purchases(c *gin.Context) {
	from, to := parseDateRange(c)
	report, err := ctrl.service.Purchases(from, to)
	if err != nil {
		response.InternalError(c, "Failed to build purchase report")
		return
	}
	response.Success(c, "Purchase report", report)
}

func (ctrl *ReportController) CustomerDue(c *gin.Context) {
	report, err := ctrl.service.CustomerDue()
	if err != nil {
		response.InternalError(c, "Failed to build customer due report")
		return
	}
	response.Success(c, "Customer due report", report)
}

func (ctrl *ReportController) SupplierDue(c *gin.Context) {
	report, err := ctrl.service.SupplierDue()
	if err != nil {
		response.InternalError(c, "Failed to build supplier due report")
		return
	}
	response.Success(c, "Supplier due report", report)
}

func (ctrl *ReportController) Stock(c *gin.Context) {
	report, err := ctrl.service.Stock()
	if err != nil {
		response.InternalError(c, "Failed to build stock report")
		return
	}
	response.Success(c, "Stock report", report)
}
