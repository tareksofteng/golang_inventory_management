package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

type PurchaseRepository interface {
	// Create runs the WHOLE purchase as one DB transaction.
	Create(purchase *models.Purchase) error
	CountAll() (int64, error)
	FindAll(search string, offset, limit int) ([]models.Purchase, int64, error)
	FindByID(id uint) (*models.Purchase, error)
	Delete(id uint) error
}

type purchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) PurchaseRepository {
	return &purchaseRepository{db: db}
}

// Create is the heart of the module. Everything inside the callback either
// fully commits or fully rolls back — if increasing stock for item #3 fails,
// the invoice insert and items #1-2 are undone too. This is what guarantees
// the books never disagree with the shelf.
//
// Laravel equivalent: DB::transaction(function () { ... }).
func (r *purchaseRepository) Create(p *models.Purchase) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Insert the invoice + its items (GORM inserts the associated
		//    Items slice automatically in the same transaction).
		if err := tx.Create(p).Error; err != nil {
			return err
		}

		// 2. Increase each product's stock and record the latest cost price.
		//    gorm.Expr keeps the increment atomic at the SQL level
		//    (quantity = quantity + ?), avoiding read-modify-write races.
		for _, item := range p.Items {
			err := tx.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				Updates(map[string]interface{}{
					"quantity":   gorm.Expr("quantity + ?", item.Quantity),
					"cost_price": item.UnitCost,
				}).Error
			if err != nil {
				return err
			}
		}

		// 3. Add the unpaid amount to the supplier's running due.
		if p.Due > 0 {
			err := tx.Model(&models.Supplier{}).
				Where("id = ?", p.SupplierID).
				Update("due", gorm.Expr("due + ?", p.Due)).Error
			if err != nil {
				return err
			}
		}

		return nil // commit
	})
}

func (r *purchaseRepository) CountAll() (int64, error) {
	var n int64
	err := r.db.Model(&models.Purchase{}).Count(&n).Error
	return n, err
}

func (r *purchaseRepository) FindAll(search string, offset, limit int) ([]models.Purchase, int64, error) {
	var purchases []models.Purchase
	var total int64

	query := r.db.Model(&models.Purchase{})
	if search != "" {
		query = query.Where("invoice_no LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Supplier").
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&purchases).Error
	if err != nil {
		return nil, 0, err
	}
	return purchases, total, nil
}

func (r *purchaseRepository) FindByID(id uint) (*models.Purchase, error) {
	var purchase models.Purchase
	err := r.db.
		Preload("Supplier").
		Preload("Items.Product").
		First(&purchase, id).Error
	if err != nil {
		return nil, err
	}
	return &purchase, nil
}

// Delete VOIDS a purchase: in one transaction it reverses the stock it added
// (clamped so it can't go negative) and the supplier due it created, then
// soft-deletes the invoice + its items.
func (r *purchaseRepository) Delete(id uint) error {
	var p models.Purchase
	if err := r.db.Preload("Items").First(&p, id).Error; err != nil {
		return err
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, it := range p.Items {
			err := tx.Model(&models.Product{}).Where("id = ?", it.ProductID).
				Update("quantity", gorm.Expr("GREATEST(quantity - ?, 0)", it.Quantity)).Error
			if err != nil {
				return err
			}
		}
		if p.Due > 0 {
			err := tx.Model(&models.Supplier{}).Where("id = ?", p.SupplierID).
				Update("due", gorm.Expr("GREATEST(due - ?, 0)", p.Due)).Error
			if err != nil {
				return err
			}
		}
		if err := tx.Where("purchase_id = ?", p.ID).Delete(&models.PurchaseItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Purchase{}, p.ID).Error
	})
}
