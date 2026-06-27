package repositories

import (
	"time"

	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// ReportRepository runs the read queries behind the Reports screens.
type ReportRepository interface {
	SalesBetween(from, to time.Time) ([]models.Sale, error)
	PurchasesBetween(from, to time.Time) ([]models.Purchase, error)
	CustomersWithDue() ([]models.Customer, error)
	SuppliersWithDue() ([]models.Supplier, error)
	StockReport() ([]models.Product, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

// SalesBetween returns sales in [from, to) with their customer preloaded.
func (r *reportRepository) SalesBetween(from, to time.Time) ([]models.Sale, error) {
	var sales []models.Sale
	err := r.db.
		Preload("Customer").
		Where("created_at >= ? AND created_at < ?", from, to).
		Order("id DESC").
		Find(&sales).Error
	return sales, err
}

func (r *reportRepository) PurchasesBetween(from, to time.Time) ([]models.Purchase, error) {
	var purchases []models.Purchase
	err := r.db.
		Preload("Supplier").
		Where("created_at >= ? AND created_at < ?", from, to).
		Order("id DESC").
		Find(&purchases).Error
	return purchases, err
}

// CustomersWithDue lists customers who owe money, biggest first.
func (r *reportRepository) CustomersWithDue() ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Where("due > 0").Order("due DESC").Find(&customers).Error
	return customers, err
}

// SuppliersWithDue lists suppliers we owe money to, biggest first.
func (r *reportRepository) SuppliersWithDue() ([]models.Supplier, error) {
	var suppliers []models.Supplier
	err := r.db.Where("due > 0").Order("due DESC").Find(&suppliers).Error
	return suppliers, err
}

// StockReport lists active products with their category for the stock sheet.
func (r *reportRepository) StockReport() ([]models.Product, error) {
	var products []models.Product
	err := r.db.
		Preload("Category").
		Where("is_active = ?", true).
		Order("quantity ASC").
		Find(&products).Error
	return products, err
}
