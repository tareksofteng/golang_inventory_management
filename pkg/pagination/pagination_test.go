package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// Parse is deliberately forgiving: bad input falls back to defaults instead of
// returning 400. The one thing it must not be forgiving about is per_page —
// an unbounded page size is a memory-exhaustion vector.
func TestParse(t *testing.T) {
	cases := []struct {
		name        string
		query       string
		wantPage    int
		wantPerPage int
		wantSearch  string
	}{
		{"no query uses defaults", "", 1, 10, ""},
		{"explicit values", "?page=3&per_page=25", 3, 25, ""},
		{"search term is passed through", "?search=phone", 1, 10, "phone"},
		{"page zero falls back to the first page", "?page=0", 1, 10, ""},
		{"negative page falls back", "?page=-4", 1, 10, ""},
		{"non-numeric page falls back", "?page=abc", 1, 10, ""},
		{"per_page above the ceiling is clamped", "?per_page=100000", 1, 100, ""},
		{"per_page at the ceiling is kept", "?per_page=100", 1, 100, ""},
		{"per_page zero falls back", "?per_page=0", 1, 10, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Parse(testContext(tc.query))

			if got.Page != tc.wantPage {
				t.Errorf("Page = %d, want %d", got.Page, tc.wantPage)
			}
			if got.PerPage != tc.wantPerPage {
				t.Errorf("PerPage = %d, want %d", got.PerPage, tc.wantPerPage)
			}
			if got.Search != tc.wantSearch {
				t.Errorf("Search = %q, want %q", got.Search, tc.wantSearch)
			}
			if got.PerPage > maxPerPage {
				t.Errorf("PerPage %d exceeds the ceiling %d", got.PerPage, maxPerPage)
			}
		})
	}
}

// Offset feeds straight into SQL OFFSET, so an off-by-one here silently skips
// or repeats a row on every list endpoint.
func TestOffset(t *testing.T) {
	cases := []struct {
		page, perPage, want int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 25, 50},
		{1, 100, 0},
	}

	for _, tc := range cases {
		p := Params{Page: tc.page, PerPage: tc.perPage}
		if got := p.Offset(); got != tc.want {
			t.Errorf("Offset(page=%d, per_page=%d) = %d, want %d", tc.page, tc.perPage, got, tc.want)
		}
	}
}

// Ceil division: a partial last page still counts as a page.
func TestTotalPages(t *testing.T) {
	cases := []struct {
		total   int64
		perPage int
		want    int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{99, 10, 10},
		{100, 10, 10},
		{101, 10, 11},
		{50, 0, 0}, // guard against division by zero
		{50, -5, 0},
	}

	for _, tc := range cases {
		if got := TotalPages(tc.total, tc.perPage); got != tc.want {
			t.Errorf("TotalPages(%d, %d) = %d, want %d", tc.total, tc.perPage, got, tc.want)
		}
	}
}

func testContext(query string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/items"+query, nil)
	return c
}
