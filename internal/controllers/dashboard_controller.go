package controllers

import (
	"inventory-api/internal/services"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	service services.DashboardService
}

func NewDashboardController(service services.DashboardService) *DashboardController {
	return &DashboardController{service: service}
}

// Summary godoc
// @Summary  Dashboard summary (KPIs, finance, low stock, top products, trend)
// @Tags     Dashboard
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  map[string]interface{}
// @Router   /dashboard/summary [get]
func (ctrl *DashboardController) Summary(c *gin.Context) {
	summary, err := ctrl.service.Summary()
	if err != nil {
		response.InternalError(c, "Failed to load dashboard")
		return
	}
	response.Success(c, "Dashboard summary fetched", summary)
}
