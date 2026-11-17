package services

import (
	"time"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"
)

// ReportSummary is the totals block shown above a sales/purchase report.
type ReportSummary struct {
	Count int     `json:"count"`
	Total float64 `json:"total"`
	Paid  float64 `json:"paid"`
	Due   float64 `json:"due"`
}

type SalesReport struct {
	From    string        `json:"from"`
	To      string        `json:"to"`
	Summary ReportSummary `json:"summary"`
	Items   []models.Sale `json:"items"`
}

type PurchaseReport struct {
	From    string            `json:"from"`
	To      string            `json:"to"`
	Summary ReportSummary     `json:"summary"`
	Items   []models.Purchase `json:"items"`
}

type DueReport struct {
	TotalDue  float64           `json:"total_due"`
	Customers []models.Customer `json:"customers,omitempty"`
	Suppliers []models.Supplier `json:"suppliers,omitempty"`
}

type StockReport struct {
	TotalValue float64          `json:"total_value"`
	Items      []models.Product `json:"items"`
}

// ProductProfitItem is one row of the product profitability report.
type ProductProfitItem struct {
	ProductID    uint    `json:"product_id"`
	Name         string  `json:"name"`
	QuantitySold int64   `json:"quantity_sold"`
	Revenue      float64 `json:"revenue"`
	Cogs         float64 `json:"cogs"`
	Profit       float64 `json:"profit"`
	Margin       float64 `json:"margin"` // profit as a percentage of revenue
}

type ProductProfitSummary struct {
	Revenue float64 `json:"revenue"`
	Cogs    float64 `json:"cogs"`
	Profit  float64 `json:"profit"`
}

type ProductProfitReport struct {
	From    string               `json:"from"`
	To      string               `json:"to"`
	Summary ProductProfitSummary `json:"summary"`
	Items   []ProductProfitItem  `json:"items"`
}

// AgingRow is one customer's outstanding balance split by how overdue it is.
type AgingRow struct {
	CustomerID uint    `json:"customer_id"`
	Customer   string  `json:"customer"`
	Current    float64 `json:"current"` // 0–30 days old
	Days30     float64 `json:"days_30"` // 31–60 days
	Days60     float64 `json:"days_60"` // 61–90 days
	Days90     float64 `json:"days_90"` // over 90 days
	Total      float64 `json:"total"`
}

// AgingReport is a receivables aging / Offene-Posten-Liste: per-customer rows
// plus a totals row, as of a given date.
type AgingReport struct {
	AsOf   string     `json:"as_of"`
	Rows   []AgingRow `json:"rows"`
	Totals AgingRow   `json:"totals"`
}

type ReportService interface {
	Sales(from, to time.Time) (*SalesReport, error)
	Purchases(from, to time.Time) (*PurchaseReport, error)
	CustomerDue() (*DueReport, error)
	SupplierDue() (*DueReport, error)
	Stock(categoryID uint) (*StockReport, error)
	ProductProfit(from, to time.Time) (*ProductProfitReport, error)
	ReceivableAging(asOf time.Time) (*AgingReport, error)
}

type reportService struct {
	repo repositories.ReportRepository
}

func NewReportService(repo repositories.ReportRepository) ReportService {
	return &reportService{repo: repo}
}

const dateLayout = "2006-01-02"

func (s *reportService) Sales(from, to time.Time) (*SalesReport, error) {
	items, err := s.repo.SalesBetween(from, to)
	if err != nil {
		return nil, err
	}
	sum := ReportSummary{Count: len(items)}
	for _, it := range items {
		sum.Total += it.TotalAmount
		sum.Paid += it.PaidAmount
		sum.Due += it.Due
	}
	return &SalesReport{From: from.Format(dateLayout), To: to.AddDate(0, 0, -1).Format(dateLayout), Summary: sum, Items: items}, nil
}

// agingBucket classifies an invoice that is `days` old into one of four aging
// buckets (0 = current 0–30, 1 = 31–60, 2 = 61–90, 3 = over 90). The boundaries
// are inclusive on the upper end, matching how a 30-day payment term reads.
func agingBucket(days int) int {
	switch {
	case days <= 30:
		return 0
	case days <= 60:
		return 1
	case days <= 90:
		return 2
	default:
		return 3
	}
}

// add places an amount into the given bucket of a row and its running total.
func (row *AgingRow) add(bucket int, amount float64) {
	switch bucket {
	case 0:
		row.Current += amount
	case 1:
		row.Days30 += amount
	case 2:
		row.Days60 += amount
	default:
		row.Days90 += amount
	}
	row.Total += amount
}

// ReceivableAging builds a per-customer aging of all outstanding sales as of the
// given date. Each unpaid invoice's balance is placed in a bucket by its age.
func (s *reportService) ReceivableAging(asOf time.Time) (*AgingReport, error) {
	sales, err := s.repo.UnpaidSales()
	if err != nil {
		return nil, err
	}

	byCustomer := make(map[uint]*AgingRow)
	order := make([]uint, 0)
	report := &AgingReport{AsOf: asOf.Format(dateLayout)}

	for _, sale := range sales {
		days := int(asOf.Sub(sale.CreatedAt).Hours() / 24)
		if days < 0 {
			days = 0 // a future-dated invoice is treated as current
		}
		bucket := agingBucket(days)

		row, ok := byCustomer[sale.CustomerID]
		if !ok {
			name := ""
			if sale.Customer != nil {
				name = sale.Customer.Name
			}
			row = &AgingRow{CustomerID: sale.CustomerID, Customer: name}
			byCustomer[sale.CustomerID] = row
			order = append(order, sale.CustomerID)
		}
		row.add(bucket, sale.Due)
		report.Totals.add(bucket, sale.Due)
	}

	for _, id := range order {
		report.Rows = append(report.Rows, *byCustomer[id])
	}
	return report, nil
}

func (s *reportService) ProductProfit(from, to time.Time) (*ProductProfitReport, error) {
	rows, err := s.repo.ProductProfit(from, to)
	if err != nil {
		return nil, err
	}

	var sum ProductProfitSummary
	items := make([]ProductProfitItem, 0, len(rows))
	for _, r := range rows {
		profit := r.Revenue - r.Cogs
		var margin float64
		if r.Revenue > 0 {
			margin = profit / r.Revenue * 100
		}
		items = append(items, ProductProfitItem{
			ProductID: r.ProductID, Name: r.Name, QuantitySold: r.QuantitySold,
			Revenue: r.Revenue, Cogs: r.Cogs, Profit: profit, Margin: margin,
		})
		sum.Revenue += r.Revenue
		sum.Cogs += r.Cogs
		sum.Profit += profit
	}

	return &ProductProfitReport{
		From:    from.Format(dateLayout),
		To:      to.AddDate(0, 0, -1).Format(dateLayout),
		Summary: sum,
		Items:   items,
	}, nil
}

func (s *reportService) Purchases(from, to time.Time) (*PurchaseReport, error) {
	items, err := s.repo.PurchasesBetween(from, to)
	if err != nil {
		return nil, err
	}
	sum := ReportSummary{Count: len(items)}
	for _, it := range items {
		sum.Total += it.TotalAmount
		sum.Paid += it.PaidAmount
		sum.Due += it.Due
	}
	return &PurchaseReport{From: from.Format(dateLayout), To: to.AddDate(0, 0, -1).Format(dateLayout), Summary: sum, Items: items}, nil
}

func (s *reportService) CustomerDue() (*DueReport, error) {
	customers, err := s.repo.CustomersWithDue()
	if err != nil {
		return nil, err
	}
	var total float64
	for _, c := range customers {
		total += c.Due
	}
	return &DueReport{TotalDue: total, Customers: customers}, nil
}

func (s *reportService) SupplierDue() (*DueReport, error) {
	suppliers, err := s.repo.SuppliersWithDue()
	if err != nil {
		return nil, err
	}
	var total float64
	for _, sup := range suppliers {
		total += sup.Due
	}
	return &DueReport{TotalDue: total, Suppliers: suppliers}, nil
}

func (s *reportService) Stock(categoryID uint) (*StockReport, error) {
	items, err := s.repo.StockReport(categoryID)
	if err != nil {
		return nil, err
	}
	var total float64
	for _, p := range items {
		total += float64(p.Quantity) * p.CostPrice
	}
	return &StockReport{TotalValue: total, Items: items}, nil
}
