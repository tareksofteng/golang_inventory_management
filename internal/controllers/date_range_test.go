package controllers

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// parseOptionalDateRange backs the date filters on the journal, sales and
// purchases lists. Two behaviours are easy to get wrong and expensive when they
// are: an absent bound must stay OPEN (zero time) rather than defaulting to
// "today" or "this month", and the `to` bound must be EXCLUSIVE of the next day
// so that a `created_at < to` query still includes everything on the end date
// itself. These tests pin both, plus the rule that malformed dates are ignored.
func TestParseOptionalDateRange(t *testing.T) {
	loc := time.Now().Location()
	day := func(s string) time.Time {
		t.Helper()
		parsed, err := time.ParseInLocation("2006-01-02", s, loc)
		if err != nil {
			t.Fatalf("bad test date %q: %v", s, err)
		}
		return parsed
	}

	cases := []struct {
		name     string
		from, to string    // raw query values ("" == param omitted)
		wantFrom time.Time // zero == unbounded
		wantTo   time.Time // zero == unbounded
	}{
		{
			name: "no params leaves both bounds open",
			from: "", to: "",
			wantFrom: time.Time{}, wantTo: time.Time{},
		},
		{
			name: "from only, to stays open",
			from: "2026-07-01", to: "",
			wantFrom: day("2026-07-01"), wantTo: time.Time{},
		},
		{
			name: "to is made exclusive by advancing one day",
			from: "", to: "2026-07-31",
			wantFrom: time.Time{}, wantTo: day("2026-08-01"),
		},
		{
			name: "both bounds set",
			from: "2026-07-01", to: "2026-07-31",
			wantFrom: day("2026-07-01"), wantTo: day("2026-08-01"),
		},
		{
			name: "single day spans from that day to the next",
			from: "2026-07-07", to: "2026-07-07",
			wantFrom: day("2026-07-07"), wantTo: day("2026-07-08"),
		},
		{
			name: "malformed from is ignored",
			from: "not-a-date", to: "",
			wantFrom: time.Time{}, wantTo: time.Time{},
		},
		{
			name: "malformed to is ignored",
			from: "", to: "07/31/2026",
			wantFrom: time.Time{}, wantTo: time.Time{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotFrom, gotTo := parseOptionalDateRange(newQueryContext(tc.from, tc.to))

			if !gotFrom.Equal(tc.wantFrom) {
				t.Errorf("from = %v, want %v", gotFrom, tc.wantFrom)
			}
			if !gotTo.Equal(tc.wantTo) {
				t.Errorf("to = %v, want %v", gotTo, tc.wantTo)
			}
		})
	}
}

// newQueryContext builds a gin.Context carrying only the from/to query params we
// care about; an empty value means the param is omitted entirely.
func newQueryContext(from, to string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	q := url.Values{}
	if from != "" {
		q.Set("from", from)
	}
	if to != "" {
		q.Set("to", to)
	}
	c.Request = httptest.NewRequest("GET", "/?"+q.Encode(), nil)
	return c
}
