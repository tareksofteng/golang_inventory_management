package services

import (
	"testing"
	"time"

	"inventory-api/internal/repositories"
)

// stubReportRepo returns canned ProductProfit rows; every other ReportRepository
// method is unreachable in these tests (and would panic via the embedded nil).
type stubReportRepo struct {
	repositories.ReportRepository
	rows []repositories.ProductProfitRow
}

func (s stubReportRepo) ProductProfit(_, _ time.Time) ([]repositories.ProductProfitRow, error) {
	return s.rows, nil
}

// The report service turns raw revenue/COGS rows into per-product profit and
// margin plus a summary. Margin is the easy one to get wrong: it must be profit
// over revenue, and must never divide by a zero revenue.
func TestProductProfitReport(t *testing.T) {
	repo := stubReportRepo{rows: []repositories.ProductProfitRow{
		{ProductID: 1, Name: "Phone", QuantitySold: 10, Revenue: 1000, Cogs: 600},
		{ProductID: 2, Name: "Loss-leader", QuantitySold: 5, Revenue: 500, Cogs: 800},
		{ProductID: 3, Name: "Freebie", QuantitySold: 2, Revenue: 0, Cogs: 0},
	}}
	svc := NewReportService(repo)

	report, err := svc.ProductProfit(time.Now(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Items) != 3 {
		t.Fatalf("got %d items, want 3", len(report.Items))
	}

	// Per-product profit = revenue - cogs; margin = profit / revenue * 100.
	assertMoney(t, "phone profit", report.Items[0].Profit, 400)
	assertMoney(t, "phone margin", report.Items[0].Margin, 40)

	// A product sold below cost shows a negative profit and margin.
	assertMoney(t, "loss profit", report.Items[1].Profit, -300)
	assertMoney(t, "loss margin", report.Items[1].Margin, -60)

	// Zero revenue must not divide by zero — margin stays 0.
	assertMoney(t, "freebie margin", report.Items[2].Margin, 0)

	// Summary totals the rows.
	assertMoney(t, "summary revenue", report.Summary.Revenue, 1500)
	assertMoney(t, "summary cogs", report.Summary.Cogs, 1400)
	assertMoney(t, "summary profit", report.Summary.Profit, 100)
}
