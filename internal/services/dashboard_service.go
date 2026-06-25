package services

import (
	"inventory-api/internal/models"
	"inventory-api/internal/repositories"
)

// lowStockThreshold: products with quantity at or below this are "low stock".
const lowStockThreshold = 10

// DashboardSummary is the full payload the dashboard screen needs in ONE call,
// so the frontend makes a single request instead of six.
type DashboardSummary struct {
	Totals             Totals                       `json:"totals"`
	StockValue         float64                      `json:"stock_value"`
	LowStockCount      int64                        `json:"low_stock_count"`
	LowStockProducts   []models.Product             `json:"low_stock_products"`
	ProductsByCategory []repositories.CategoryCount `json:"products_by_category"`
	RecentProducts     []models.Product             `json:"recent_products"`
}

type Totals struct {
	Products   int64 `json:"products"`
	Categories int64 `json:"categories"`
	Suppliers  int64 `json:"suppliers"`
	Customers  int64 `json:"customers"`
}

type DashboardService interface {
	Summary() (*DashboardSummary, error)
}

type dashboardService struct {
	repo repositories.DashboardRepository
}

func NewDashboardService(repo repositories.DashboardRepository) DashboardService {
	return &dashboardService{repo: repo}
}

// Summary assembles every dashboard metric. Each step is a cheap aggregate
// query; we surface the first error rather than returning a half-built page.
func (s *dashboardService) Summary() (*DashboardSummary, error) {
	var (
		out DashboardSummary
		err error
	)

	if out.Totals.Products, err = s.repo.Count(&models.Product{}); err != nil {
		return nil, err
	}
	if out.Totals.Categories, err = s.repo.Count(&models.Category{}); err != nil {
		return nil, err
	}
	if out.Totals.Suppliers, err = s.repo.Count(&models.Supplier{}); err != nil {
		return nil, err
	}
	if out.Totals.Customers, err = s.repo.Count(&models.Customer{}); err != nil {
		return nil, err
	}

	if out.StockValue, err = s.repo.TotalStockValue(); err != nil {
		return nil, err
	}

	if out.LowStockProducts, out.LowStockCount, err = s.repo.LowStockProducts(lowStockThreshold, 10); err != nil {
		return nil, err
	}

	if out.ProductsByCategory, err = s.repo.ProductsByCategory(); err != nil {
		return nil, err
	}

	if out.RecentProducts, err = s.repo.RecentProducts(5); err != nil {
		return nil, err
	}

	return &out, nil
}
