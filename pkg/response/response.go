package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Every JSON response the API returns uses this single envelope shape so the
// frontend can rely on one predictable structure. Compare with Laravel API
// Resources / response()->json(); here we own the contract explicitly.
//
// Example success:
//
//	{ "success": true,  "message": "Category created", "data": {...} }
//
// Example error:
//
//	{ "success": false, "message": "Validation failed", "errors": {...} }
type envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Meta carries pagination information. Pointer + omitempty means it only
// appears in the JSON for list endpoints, never for single-item responses.
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Success sends a 200 OK with a data payload.
func Success(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 Created — use after a successful POST.
func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Paginated sends a 200 OK with both a data list and pagination meta.
func Paginated(c *gin.Context, message string, data any, meta Meta) {
	c.JSON(http.StatusOK, envelope{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

// Error sends a failure response with any HTTP status code. The errs argument
// is optional detail (e.g. a validation field map); pass nil when not needed.
func Error(c *gin.Context, status int, message string, errs any) {
	c.JSON(status, envelope{
		Success: false,
		Message: message,
		Errors:  errs,
	})
}

// ---- Thin convenience wrappers around Error for the common cases ----

// BadRequest -> 400. Typically used for validation failures.
func BadRequest(c *gin.Context, message string, errs any) {
	Error(c, http.StatusBadRequest, message, errs)
}

// NotFound -> 404.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, nil)
}

// InternalError -> 500. Use for unexpected server-side failures.
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message, nil)
}
