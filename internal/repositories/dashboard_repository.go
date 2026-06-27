package repositories

import (
	"time"

	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// CategoryCount is one row of the "products per category" breakdown (for a chart).
type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// ProductSold is one row of the "top selling products" report.
type ProductSold struct {
	ProductID    uint    `json:"product_id"`
	Name         string  `json:"name"`
	QuantitySold int64   `json:"quantity_sold"`
	Revenue      float64 `json:"revenue"`
}

// DayTotal is one day's sales total (for the trend chart).
type DayTotal struct {
	Date  string  `json:"date"`
	Total float64 `json:"total"`
}

// DashboardRepository runs the read-only aggregate queries the dashboard needs.
type DashboardRepository interface {
	Count(model interface{}) (int64, error)
	TotalStockValue() (float64, error)
	LowStockProducts(threshold, limit int) ([]models.Product, int64, error)
	ProductsByCategory() ([]CategoryCount, error)

	// Sales / purchase money sums. A nil `from` means "all time".
	SalesSum(from *time.Time) (float64, error)
	PurchaseSum(from *time.Time) (float64, error)

	// Outstanding balances.
	CustomerDueTotal() (float64, error)
	SupplierDueTotal() (float64, error)

	TopSellingProducts(limit int) ([]ProductSold, error)
	SalesByDay(from time.Time) ([]DayTotal, error)
	RecentSales(limit int) ([]models.Sale, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) Count(model interface{}) (int64, error) {
	var n int64
	err := r.db.Model(model).Count(&n).Error
	return n, err
}

func (r *dashboardRepository) TotalStockValue() (float64, error) {
	var value float64
	err := r.db.Model(&models.Product{}).
		Select("COALESCE(SUM(quantity * cost_price), 0)").
		Scan(&value).Error
	return value, err
}

func (r *dashboardRepository) LowStockProducts(threshold, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	base := r.db.Model(&models.Product{}).
		Where("is_active = ? AND quantity <= ?", true, threshold)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := base.Preload("Category").Order("quantity ASC").Limit(limit).Find(&products).Error
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *dashboardRepository) ProductsByCategory() ([]CategoryCount, error) {
	var rows []CategoryCount
	err := r.db.Table("products").
		Select("categories.name as category, COUNT(products.id) as count").
		Joins("JOIN categories ON categories.id = products.category_id").
		Where("products.deleted_at IS NULL").
		Group("categories.id, categories.name").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

// SalesSum sums sales.total_amount, optionally from a start time.
func (r *dashboardRepository) SalesSum(from *time.Time) (float64, error) {
	return r.sumAmount(&models.Sale{}, from)
}

func (r *dashboardRepository) PurchaseSum(from *time.Time) (float64, error) {
	return r.sumAmount(&models.Purchase{}, from)
}

// sumAmount is the shared SUM(total_amount) helper for sales & purchases.
func (r *dashboardRepository) sumAmount(model interface{}, from *time.Time) (float64, error) {
	var value float64
	q := r.db.Model(model).Select("COALESCE(SUM(total_amount), 0)")
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	err := q.Scan(&value).Error
	return value, err
}

func (r *dashboardRepository) CustomerDueTotal() (float64, error) {
	var v float64
	err := r.db.Model(&models.Customer{}).Select("COALESCE(SUM(due), 0)").Scan(&v).Error
	return v, err
}

func (r *dashboardRepository) SupplierDueTotal() (float64, error) {
	var v float64
	err := r.db.Model(&models.Supplier{}).Select("COALESCE(SUM(due), 0)").Scan(&v).Error
	return v, err
}

// TopSellingProducts ranks products by total units sold.
func (r *dashboardRepository) TopSellingProducts(limit int) ([]ProductSold, error) {
	var rows []ProductSold
	err := r.db.Table("sale_items").
		Select("products.id as product_id, products.name as name, SUM(sale_items.quantity) as quantity_sold, SUM(sale_items.subtotal) as revenue").
		Joins("JOIN products ON products.id = sale_items.product_id").
		Where("sale_items.deleted_at IS NULL").
		Group("products.id, products.name").
		Order("quantity_sold DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// SalesByDay returns daily sales totals from a start date (for the trend chart).
func (r *dashboardRepository) SalesByDay(from time.Time) ([]DayTotal, error) {
	var rows []DayTotal
	err := r.db.Table("sales").
		Select("DATE(created_at) as date, COALESCE(SUM(total_amount), 0) as total").
		Where("created_at >= ? AND deleted_at IS NULL", from).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *dashboardRepository) RecentSales(limit int) ([]models.Sale, error) {
	var sales []models.Sale
	err := r.db.Preload("Customer").Order("id DESC").Limit(limit).Find(&sales).Error
	return sales, err
}
