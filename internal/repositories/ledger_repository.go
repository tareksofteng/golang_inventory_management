package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// LedgerRepository fetches every transaction that moves a party's balance, so
// the service can weave them into a running-balance statement.
type LedgerRepository interface {
	SalesByCustomer(customerID uint) ([]models.Sale, error)
	SaleReturnsByCustomer(customerID uint) ([]models.SaleReturn, error)
	PurchasesBySupplier(supplierID uint) ([]models.Purchase, error)
	PurchaseReturnsBySupplier(supplierID uint) ([]models.PurchaseReturn, error)
	PaymentsByParty(partyType string, partyID uint) ([]models.Payment, error)
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
