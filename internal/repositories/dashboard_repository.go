package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// CategoryCount is one row of the "products per category" breakdown (for a chart).
type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// DashboardRepository runs the read-only aggregate queries the dashboard needs.
type DashboardRepository interface {
	Count(model interface{}) (int64, error)
	TotalStockValue() (float64, error)
	LowStockProducts(threshold, limit int) ([]models.Product, int64, error)
	ProductsByCategory() ([]CategoryCount, error)
	RecentProducts(limit int) ([]models.Product, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

// Count returns the (soft-delete-aware) row count for any model, e.g.
// Count(&models.Product{}). One helper avoids four near-identical methods.
func (r *dashboardRepository) Count(model interface{}) (int64, error) {
	var n int64
	err := r.db.Model(model).Count(&n).Error
	return n, err
}

// TotalStockValue = SUM(quantity * cost_price) across active products.
func (r *dashboardRepository) TotalStockValue() (float64, error) {
	var value float64
	err := r.db.Model(&models.Product{}).
		Select("COALESCE(SUM(quantity * cost_price), 0)").
		Scan(&value).Error
	return value, err
}

// LowStockProducts returns active products at/under the threshold (with their
// Category preloaded) plus the total low-stock count.
func (r *dashboardRepository) LowStockProducts(threshold, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	base := r.db.Model(&models.Product{}).
		Where("is_active = ? AND quantity <= ?", true, threshold)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := base.
		Preload("Category").
		Order("quantity ASC").
		Limit(limit).
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

// ProductsByCategory groups product counts by category name (chart data).
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

// RecentProducts returns the latest-created products with their relations.
func (r *dashboardRepository) RecentProducts(limit int) ([]models.Product, error) {
	var products []models.Product
	err := r.db.
		Preload("Category").
		Preload("Supplier").
		Order("id DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}
