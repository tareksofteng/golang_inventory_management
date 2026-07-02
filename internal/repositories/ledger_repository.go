package repositories

import (
	"time"

	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// Movement is one stock-affecting line for a product (quantity in or out).
type Movement struct {
	Date time.Time
	Ref  string
	Qty  int
}

// LedgerRepository fetches every transaction that moves a party's balance, so
// the service can weave them into a running-balance statement.
type LedgerRepository interface {
	SalesByCustomer(customerID uint) ([]models.Sale, error)
	SaleReturnsByCustomer(customerID uint) ([]models.SaleReturn, error)
	PurchasesBySupplier(supplierID uint) ([]models.Purchase, error)
	PurchaseReturnsBySupplier(supplierID uint) ([]models.PurchaseReturn, error)
	PaymentsByParty(partyType string, partyID uint) ([]models.Payment, error)

	// Product stock movements (for the product ledger).
	ProductPurchases(productID uint) ([]Movement, error)       // stock IN
	ProductSales(productID uint) ([]Movement, error)           // stock OUT
	ProductPurchaseReturns(productID uint) ([]Movement, error) // stock OUT
	ProductSaleReturns(productID uint) ([]Movement, error)     // stock IN
}

type ledgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) LedgerRepository {
	return &ledgerRepository{db: db}
}

func (r *ledgerRepository) SalesByCustomer(customerID uint) ([]models.Sale, error) {
	var rows []models.Sale
	err := r.db.Where("customer_id = ?", customerID).Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *ledgerRepository) SaleReturnsByCustomer(customerID uint) ([]models.SaleReturn, error) {
	var rows []models.SaleReturn
	err := r.db.Where("customer_id = ?", customerID).Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *ledgerRepository) PurchasesBySupplier(supplierID uint) ([]models.Purchase, error) {
	var rows []models.Purchase
	err := r.db.Where("supplier_id = ?", supplierID).Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *ledgerRepository) PurchaseReturnsBySupplier(supplierID uint) ([]models.PurchaseReturn, error) {
	var rows []models.PurchaseReturn
	err := r.db.Where("supplier_id = ?", supplierID).Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *ledgerRepository) PaymentsByParty(partyType string, partyID uint) ([]models.Payment, error) {
	var rows []models.Payment
	err := r.db.Where("party_type = ? AND party_id = ?", partyType, partyID).Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *ledgerRepository) ProductPurchases(productID uint) ([]Movement, error) {
	var rows []Movement
	err := r.db.Table("purchase_items").
		Select("purchases.created_at as date, purchases.invoice_no as ref, purchase_items.quantity as qty").
		Joins("JOIN purchases ON purchases.id = purchase_items.purchase_id").
		Where("purchase_items.product_id = ? AND purchase_items.deleted_at IS NULL AND purchases.deleted_at IS NULL", productID).
		Scan(&rows).Error
	return rows, err
}

func (r *ledgerRepository) ProductSales(productID uint) ([]Movement, error) {
	var rows []Movement
	err := r.db.Table("sale_items").
		Select("sales.created_at as date, sales.invoice_no as ref, sale_items.quantity as qty").
		Joins("JOIN sales ON sales.id = sale_items.sale_id").
		Where("sale_items.product_id = ? AND sale_items.deleted_at IS NULL AND sales.deleted_at IS NULL", productID).
		Scan(&rows).Error
	return rows, err
}

func (r *ledgerRepository) ProductPurchaseReturns(productID uint) ([]Movement, error) {
	var rows []Movement
	err := r.db.Table("purchase_return_items").
		Select("purchase_returns.created_at as date, purchase_returns.invoice_no as ref, purchase_return_items.quantity as qty").
		Joins("JOIN purchase_returns ON purchase_returns.id = purchase_return_items.purchase_return_id").
		Where("purchase_return_items.product_id = ? AND purchase_return_items.deleted_at IS NULL AND purchase_returns.deleted_at IS NULL", productID).
		Scan(&rows).Error
	return rows, err
}

func (r *ledgerRepository) ProductSaleReturns(productID uint) ([]Movement, error) {
	var rows []Movement
	err := r.db.Table("sale_return_items").
		Select("sale_returns.created_at as date, sale_returns.invoice_no as ref, sale_return_items.quantity as qty").
		Joins("JOIN sale_returns ON sale_returns.id = sale_return_items.sale_return_id").
		Where("sale_return_items.product_id = ? AND sale_return_items.deleted_at IS NULL AND sale_returns.deleted_at IS NULL", productID).
		Scan(&rows).Error
	return rows, err
}
