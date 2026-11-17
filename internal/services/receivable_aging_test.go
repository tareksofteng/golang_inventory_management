package services

import (
	"testing"
	"time"

	"inventory-api/internal/models"
)

// agingBucket decides which column an invoice lands in. The boundaries (30, 60,
// 90 days) are exactly where off-by-one mistakes hide, so each edge is pinned.
func TestAgingBucket(t *testing.T) {
	cases := []struct {
		days int
		want int
	}{
		{0, 0}, {15, 0}, {30, 0}, // current: up to and including 30 days
		{31, 1}, {60, 1}, // 31–60
		{61, 2}, {90, 2}, // 61–90
		{91, 3}, {365, 3}, // over 90
	}
	for _, c := range cases {
		if got := agingBucket(c.days); got != c.want {
			t.Errorf("agingBucket(%d) = %d, want %d", c.days, got, c.want)
		}
	}
}

// saleAgedDays is a small helper: an unpaid sale created `days` ago.
func saleAgedDays(asOf time.Time, customerID uint, customer string, due float64, days int) models.Sale {
	s := models.Sale{CustomerID: customerID, Due: due, Customer: &models.Customer{Name: customer}}
	s.CreatedAt = asOf.AddDate(0, 0, -days)
	return s
}

// ReceivableAging must bucket each outstanding invoice by age, aggregate the
// invoices per customer, and total every bucket across the whole report.
func TestReceivableAging(t *testing.T) {
	asOf := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

	repo := stubReportRepo{sales: []models.Sale{
		saleAgedDays(asOf, 1, "Acme GmbH", 1000, 10),  // current
		saleAgedDays(asOf, 1, "Acme GmbH", 500, 45),   // 31–60
		saleAgedDays(asOf, 2, "Berlin AG", 2000, 120), // 90+
	}}
	svc := NewReportService(repo)

	report, err := svc.ReceivableAging(asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Rows) != 2 {
		t.Fatalf("got %d customer rows, want 2", len(report.Rows))
	}

	// The two Acme invoices collapse into one row, split across two buckets.
	acme := report.Rows[0]
	if acme.Customer != "Acme GmbH" {
		t.Fatalf("expected Acme first (insertion order), got %q", acme.Customer)
	}
	assertMoney(t, "acme current", acme.Current, 1000)
	assertMoney(t, "acme 31-60", acme.Days30, 500)
	assertMoney(t, "acme total", acme.Total, 1500)

	// Berlin's single old invoice lands in the 90+ bucket.
	berlin := report.Rows[1]
	assertMoney(t, "berlin 90+", berlin.Days90, 2000)

	// Totals sum every bucket across all customers.
	assertMoney(t, "totals current", report.Totals.Current, 1000)
	assertMoney(t, "totals 31-60", report.Totals.Days30, 500)
	assertMoney(t, "totals 90+", report.Totals.Days90, 2000)
	assertMoney(t, "totals grand", report.Totals.Total, 3500)
}
