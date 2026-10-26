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

type ReportService interface {
	Sales(from, to time.Time) (*SalesReport, error)
	Purchases(from, to time.Time) (*PurchaseReport, error)
	CustomerDue() (*DueReport, error)
	SupplierDue() (*DueReport, error)
	Stock(categoryID uint) (*StockReport, error)
	ProductProfit(from, to time.Time) (*ProductProfitReport, error)
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
