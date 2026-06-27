package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

type ReturnRepository interface {
	CreatePurchaseReturn(r *models.PurchaseReturn) error
	CreateSaleReturn(r *models.SaleReturn) error
	CountPurchaseReturns() (int64, error)
	CountSaleReturns() (int64, error)
	FindPurchaseReturns(offset, limit int) ([]models.PurchaseReturn, int64, error)
	FindSaleReturns(offset, limit int) ([]models.SaleReturn, int64, error)
}

type returnRepository struct {
	db *gorm.DB
}

func NewReturnRepository(db *gorm.DB) ReturnRepository {
	return &returnRepository{db: db}
}

// CreatePurchaseReturn: insert + DECREASE stock (guarded) + reduce supplier due.
func (r *returnRepository) CreatePurchaseReturn(ret *models.PurchaseReturn) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ret).Error; err != nil {
			return err
		}
		for _, item := range ret.Items {
			res := tx.Model(&models.Product{}).
				Where("id = ? AND quantity >= ?", item.ProductID, item.Quantity).
				Update("quantity", gorm.Expr("quantity - ?", item.Quantity))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return ErrInsufficientStock // cannot return more than we hold
			}
		}
		// Reduce supplier due, never below zero.
		return tx.Model(&models.Supplier{}).
			Where("id = ?", ret.SupplierID).
			Update("due", gorm.Expr("GREATEST(due - ?, 0)", ret.TotalAmount)).Error
	})
}

// CreateSaleReturn: insert + INCREASE stock + reduce customer due.
func (r *returnRepository) CreateSaleReturn(ret *models.SaleReturn) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ret).Error; err != nil {
			return err
		}
		for _, item := range ret.Items {
			err := tx.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				Update("quantity", gorm.Expr("quantity + ?", item.Quantity)).Error
			if err != nil {
				return err
			}
		}
		return tx.Model(&models.Customer{}).
			Where("id = ?", ret.CustomerID).
			Update("due", gorm.Expr("GREATEST(due - ?, 0)", ret.TotalAmount)).Error
	})
}

func (r *returnRepository) CountPurchaseReturns() (int64, error) {
	var n int64
	err := r.db.Model(&models.PurchaseReturn{}).Count(&n).Error
	return n, err
}

func (r *returnRepository) CountSaleReturns() (int64, error) {
	var n int64
	err := r.db.Model(&models.SaleReturn{}).Count(&n).Error
	return n, err
}

func (r *returnRepository) FindPurchaseReturns(offset, limit int) ([]models.PurchaseReturn, int64, error) {
	var rows []models.PurchaseReturn
	var total int64
	if err := r.db.Model(&models.PurchaseReturn{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Preload("Supplier").Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *returnRepository) FindSaleReturns(offset, limit int) ([]models.SaleReturn, int64, error) {
	var rows []models.SaleReturn
	var total int64
	if err := r.db.Model(&models.SaleReturn{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Preload("Customer").Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}
