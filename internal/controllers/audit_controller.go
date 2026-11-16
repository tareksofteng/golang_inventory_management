package controllers

import (
	"inventory-api/internal/repositories"
	"inventory-api/pkg/pagination"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuditController struct {
	repo repositories.AuditLogRepository
}

func NewAuditController(repo repositories.AuditLogRepository) *AuditController {
	return &AuditController{repo: repo}
}

// List godoc
// @Summary  List the write-operation audit trail (paginated, newest first)
// @Tags     Audit
// @Produce  json
// @Security BearerAuth
// @Param    page      query     int  false  "Page number"
// @Param    per_page  query     int  false  "Items per page"
// @Success  200       {object}  map[string]interface{}
// @Router   /audit-logs [get]
func (ctrl *AuditController) List(c *gin.Context) {
	p := pagination.Parse(c)
	logs, total, err := ctrl.repo.List((p.Page-1)*p.PerPage, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch audit logs")
		return
	}
	response.Paginated(c, "Audit logs", logs, response.Meta{
		Page: p.Page, PerPage: p.PerPage, Total: total, TotalPages: pagination.TotalPages(total, p.PerPage),
	})
}
