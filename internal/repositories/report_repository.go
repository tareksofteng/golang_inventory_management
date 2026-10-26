package repositories

import (
	"time"

	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// ProductProfitRow is one product's sales performance over a period. Revenue is
// the sum of line subtotals (quantity * unit price); invoice-level discounts are
// not allocated per line, matching the dashboard's top-selling revenue.
type ProductProfitRow struct {
	ProductID    uint    `json:"product_id"`
	Name         string  `json:"name"`
	QuantitySold int64   `json:"quantity_sold"`
	Revenue      float64 `json:"revenue"`
	Cogs         float64 `json:"cogs"`
}

// ReportRepository runs the read queries behind the Reports screens.
type ReportRepository interface {
	SalesBetween(from, to time.Time) ([]models.Sale, error)
	PurchasesBetween(from, to time.Time) ([]models.Purchase, error)
	CustomersWithDue() ([]models.Customer, error)
	SuppliersWithDue() ([]models.Supplier, error)
	StockReport(categoryID uint) ([]models.Product, error)
	ProductProfit(from, to time.Time) ([]ProductProfitRow, error)
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

// ProductProfit aggregates sold units, revenue and COGS per product for sales
// in [from, to), ordered by gross profit (revenue - cogs) descending.
func (r *reportRepository) ProductProfit(from, to time.Time) ([]ProductProfitRow, error) {
	var rows []ProductProfitRow
	err := r.db.Table("sale_items").
		Select(
			"products.id as product_id, products.name as name, "+
				"SUM(sale_items.quantity) as quantity_sold, "+
				"SUM(sale_items.subtotal) as revenue, "+
				"SUM(sale_items.unit_cost * sale_items.quantity) as cogs").
		Joins("JOIN sales ON sales.id = sale_items.sale_id").
		Joins("JOIN products ON products.id = sale_items.product_id").
		Where("sale_items.deleted_at IS NULL AND sales.deleted_at IS NULL").
		Where("sales.created_at >= ? AND sales.created_at < ?", from, to).
		Group("products.id, products.name").
		Order("SUM(sale_items.subtotal) - SUM(sale_items.unit_cost * sale_items.quantity) DESC").
		Scan(&rows).Error
	return rows, err
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

// StockReport lists active products with their category for the stock sheet,
// optionally filtered to a single category (categoryID == 0 means all).
func (r *reportRepository) StockReport(categoryID uint) ([]models.Product, error) {
	var products []models.Product
	q := r.db.Preload("Category").Where("is_active = ?", true)
	if categoryID > 0 {
		q = q.Where("category_id = ?", categoryID)
	}
	err := q.Order("quantity ASC").Find(&products).Error
	return products, err
}
