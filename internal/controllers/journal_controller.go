package controllers

import (
	"errors"
	"net/http"
	"time"

	"inventory-api/internal/middleware"
	"inventory-api/internal/services"
	"inventory-api/pkg/pagination"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type JournalLineRequest struct {
	AccountID uint    `json:"account_id" binding:"required"`
	Debit     float64 `json:"debit" binding:"gte=0"`
	Credit    float64 `json:"credit" binding:"gte=0"`
}

type CreateJournalRequest struct {
	Date      string               `json:"date"` // YYYY-MM-DD (optional, defaults to today)
	Reference string               `json:"reference" binding:"max=100"`
	Note      string               `json:"note" binding:"max=255"`
	Lines     []JournalLineRequest `json:"lines" binding:"required,min=2,dive"`
}

type JournalController struct {
	service services.JournalService
}

func NewJournalController(service services.JournalService) *JournalController {
	return &JournalController{service: service}
}

// Create godoc
// @Summary  Post a double-entry journal entry (debit must equal credit)
// @Tags     Journal
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    body  body      CreateJournalRequest  true  "Journal entry"
// @Success  201   {object}  map[string]interface{}
// @Failure  422   {object}  map[string]interface{}
// @Router   /journal [post]
func (ctrl *JournalController) Create(c *gin.Context) {
	var req CreateJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	var date time.Time
	if req.Date != "" {
		if t, err := time.Parse("2006-01-02", req.Date); err == nil {
			date = t
		}
	}

	lines := make([]services.JournalLineInput, len(req.Lines))
	for i, l := range req.Lines {
		lines[i] = services.JournalLineInput{AccountID: l.AccountID, Debit: l.Debit, Credit: l.Credit}
	}

	entry, err := ctrl.service.Create(services.CreateJournalInput{
		Date:      date,
		Reference: req.Reference,
		Note:      req.Note,
		UserID:    middleware.UserID(c),
		Lines:     lines,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrTooFewLines),
			errors.Is(err, services.ErrInvalidLine),
			errors.Is(err, services.ErrUnbalanced),
			errors.Is(err, services.ErrAccountNotFound):
			response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			response.InternalError(c, "Failed to post journal entry")
		}
		return
	}
	response.Created(c, "Journal entry posted", entry)
}

// List godoc
// @Summary  List journal entries (paginated)
// @Tags     Journal
// @Produce  json
// @Security BearerAuth
// @Param    page      query     int  false  "Page number"
// @Param    per_page  query     int  false  "Items per page"
// @Success  200       {object}  map[string]interface{}
// @Router   /journal [get]
func (ctrl *JournalController) List(c *gin.Context) {
	p := pagination.Parse(c)
	entries, total, err := ctrl.service.List(p.Page, p.PerPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch journal entries")
		return
	}
	response.Paginated(c, "Journal entries", entries, response.Meta{
		Page: p.Page, PerPage: p.PerPage, Total: total, TotalPages: pagination.TotalPages(total, p.PerPage),
	})
}

// Get godoc
// @Summary  Get a journal entry with its lines
// @Tags     Journal
// @Produce  json
// @Security BearerAuth
// @Param    id   path      int  true  "Journal entry ID"
// @Success  200  {object}  map[string]interface{}
// @Router   /journal/{id} [get]
func (ctrl *JournalController) Get(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "Invalid id", nil)
		return
	}
	entry, err := ctrl.service.Get(id)
	if err != nil {
		if errors.Is(err, services.ErrJournalNotFound) {
			response.NotFound(c, "Journal entry not found")
			return
		}
		response.InternalError(c, "Failed to fetch journal entry")
		return
	}
	response.Success(c, "Journal entry", entry)
}
