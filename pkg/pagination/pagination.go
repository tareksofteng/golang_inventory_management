package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Sensible defaults and a hard ceiling so a client cannot ask for, say,
// 1,000,000 rows in one page and exhaust the server's memory.
const (
	defaultPage    = 1
	defaultPerPage = 10
	maxPerPage     = 100
)

// Params holds the parsed list-query inputs shared by every list endpoint:
// pagination (page, per_page) plus a free-text search term.
type Params struct {
	Page    int
	PerPage int
	Search  string
}

// Offset converts page/per_page into the SQL OFFSET value.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Parse reads ?page=, ?per_page= and ?search= from the request, applying
// defaults and clamping. Invalid/missing values fall back to defaults instead
// of erroring — a forgiving list endpoint is good UX.
//
// Example: GET /categories?page=2&per_page=20&search=phone
func Parse(c *gin.Context) Params {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = defaultPage
	}

	perPage, _ := strconv.Atoi(c.Query("per_page"))
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return Params{
		Page:    page,
		PerPage: perPage,
		Search:  c.Query("search"),
	}
}

// TotalPages computes how many pages exist for a given total row count.
// Ceil division without floats: (total + perPage - 1) / perPage.
func TotalPages(total int64, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	return int((total + int64(perPage) - 1) / int64(perPage))
}
