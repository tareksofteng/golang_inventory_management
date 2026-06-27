package services

import (
	"time"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"
)

// lowStockThreshold: products with quantity at or below this are "low stock".
const lowStockThreshold = 10

// DashboardSummary is the full payload the dashboard screen needs in ONE call.
type DashboardSummary struct {
	Totals             Totals                       `json:"totals"`
	Finance            Finance                      `json:"finance"`
	StockValue         float64                      `json:"stock_value"`
	LowStockCount      int64                        `json:"low_stock_count"`
	LowStockProducts   []models.Product             `json:"low_stock_products"`
	ProductsByCategory []repositories.CategoryCount `json:"products_by_category"`
	TopSellingProducts []repositories.ProductSold   `json:"top_selling_products"`
	SalesTrend         []repositories.DayTotal      `json:"sales_trend"`
	RecentSales        []models.Sale                `json:"recent_sales"`
}

type Totals struct {
	Products   int64 `json:"products"`
	Categories int64 `json:"categories"`
	Suppliers  int64 `json:"suppliers"`
	Customers  int64 `json:"customers"`
}

type Finance struct {
	TotalSales    float64 `json:"total_sales"`
	TotalPurchase float64 `json:"total_purchase"`
	TodaySales    float64 `json:"today_sales"`
	TodayPurchase float64 `json:"today_purchase"`
	MonthSales    float64 `json:"month_sales"`
	Receivable    float64 `json:"receivable"` // total customer due (money owed TO us)
	Payable       float64 `json:"payable"`    // total supplier due (money WE owe)
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

func (s *dashboardService) Summary() (*DashboardSummary, error) {
	var (
		out DashboardSummary
		err error
	)

	// --- Totals ---
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

	// --- Finance (time-windowed sums) ---
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	if out.Finance.TotalSales, err = s.repo.SalesSum(nil); err != nil {
		return nil, err
	}
	if out.Finance.TotalPurchase, err = s.repo.PurchaseSum(nil); err != nil {
		return nil, err
	}
	if out.Finance.TodaySales, err = s.repo.SalesSum(&todayStart); err != nil {
		return nil, err
	}
	if out.Finance.TodayPurchase, err = s.repo.PurchaseSum(&todayStart); err != nil {
		return nil, err
	}
	if out.Finance.MonthSales, err = s.repo.SalesSum(&monthStart); err != nil {
		return nil, err
	}
	if out.Finance.Receivable, err = s.repo.CustomerDueTotal(); err != nil {
		return nil, err
	}
	if out.Finance.Payable, err = s.repo.SupplierDueTotal(); err != nil {
		return nil, err
	}

	// --- Inventory widgets ---
	if out.StockValue, err = s.repo.TotalStockValue(); err != nil {
		return nil, err
	}
	if out.LowStockProducts, out.LowStockCount, err = s.repo.LowStockProducts(lowStockThreshold, 10); err != nil {
		return nil, err
	}
	if out.ProductsByCategory, err = s.repo.ProductsByCategory(); err != nil {
		return nil, err
	}
	if out.TopSellingProducts, err = s.repo.TopSellingProducts(5); err != nil {
		return nil, err
	}
	if out.RecentSales, err = s.repo.RecentSales(5); err != nil {
		return nil, err
	}

	// --- 7-day sales trend (fill gaps so the chart has all days) ---
	if out.SalesTrend, err = s.salesTrend(todayStart); err != nil {
		return nil, err
	}

	return &out, nil
}

// salesTrend builds a dense last-7-days array (one entry per day, 0 where there
// were no sales) so the frontend chart never has holes.
func (s *dashboardService) salesTrend(todayStart time.Time) ([]repositories.DayTotal, error) {
	weekStart := todayStart.AddDate(0, 0, -6)
	rows, err := s.repo.SalesByDay(weekStart)
	if err != nil {
		return nil, err
	}

	byDate := make(map[string]float64, len(rows))
	for _, r := range rows {
		// MySQL DATE() comes back as "2006-01-02" (sometimes with time); keep
		// the first 10 chars to be safe.
		key := r.Date
		if len(key) > 10 {
			key = key[:10]
		}
		byDate[key] = r.Total
	}

	trend := make([]repositories.DayTotal, 0, 7)
	for i := 0; i < 7; i++ {
		day := weekStart.AddDate(0, 0, i).Format("2006-01-02")
		trend = append(trend, repositories.DayTotal{Date: day, Total: byDate[day]})
	}
	return trend, nil
}
