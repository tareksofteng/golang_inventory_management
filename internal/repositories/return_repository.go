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
	FindPurchaseReturnByID(id uint) (*models.PurchaseReturn, error)
	FindSaleReturnByID(id uint) (*models.SaleReturn, error)
	// Already-returned quantity per product for a given source invoice.
	ReturnedQtyByPurchase(purchaseID uint) (map[uint]int, error)
	ReturnedQtyBySale(saleID uint) (map[uint]int, error)
}

type productQty struct {
	ProductID uint
	Qty       int
}

func toQtyMap(rows []productQty) map[uint]int {
	m := make(map[uint]int, len(rows))
	for _, r := range rows {
		m[r.ProductID] = r.Qty
	}
	return m
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

func (r *returnRepository) FindPurchaseReturnByID(id uint) (*models.PurchaseReturn, error) {
	var ret models.PurchaseReturn
	err := r.db.Preload("Supplier").Preload("Items.Product").First(&ret, id).Error
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (r *returnRepository) FindSaleReturnByID(id uint) (*models.SaleReturn, error) {
	var ret models.SaleReturn
	err := r.db.Preload("Customer").Preload("Items.Product").First(&ret, id).Error
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (r *returnRepository) ReturnedQtyByPurchase(purchaseID uint) (map[uint]int, error) {
	var rows []productQty
	err := r.db.Table("purchase_return_items").
		Select("purchase_return_items.product_id as product_id, SUM(purchase_return_items.quantity) as qty").
		Joins("JOIN purchase_returns ON purchase_returns.id = purchase_return_items.purchase_return_id").
		Where("purchase_returns.purchase_id = ? AND purchase_return_items.deleted_at IS NULL", purchaseID).
		Group("purchase_return_items.product_id").
		Scan(&rows).Error
	return toQtyMap(rows), err
}

func (r *returnRepository) ReturnedQtyBySale(saleID uint) (map[uint]int, error) {
	var rows []productQty
	err := r.db.Table("sale_return_items").
		Select("sale_return_items.product_id as product_id, SUM(sale_return_items.quantity) as qty").
		Joins("JOIN sale_returns ON sale_returns.id = sale_return_items.sale_return_id").
		Where("sale_returns.sale_id = ? AND sale_return_items.deleted_at IS NULL", saleID).
		Group("sale_return_items.product_id").
		Scan(&rows).Error
	return toQtyMap(rows), err
}
