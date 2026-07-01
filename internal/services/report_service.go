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

type ReportService interface {
	Sales(from, to time.Time) (*SalesReport, error)
	Purchases(from, to time.Time) (*PurchaseReport, error)
	CustomerDue() (*DueReport, error)
	SupplierDue() (*DueReport, error)
	Stock(categoryID uint) (*StockReport, error)
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
