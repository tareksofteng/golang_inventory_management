package repositories

import (
	"errors"

	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// ErrInsufficientStock is returned by Create when a line wants more units than
// are in stock. It lives here (not in services) because the transaction below
// is what detects the condition; the service translates it for the controller.
var ErrInsufficientStock = errors.New("insufficient stock")

type SaleRepository interface {
	Create(sale *models.Sale) error
	CountAll() (int64, error)
	FindAll(search string, offset, limit int) ([]models.Sale, int64, error)
	FindByID(id uint) (*models.Sale, error)
	Delete(id uint) error
}

type saleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) SaleRepository {
	return &saleRepository{db: db}
}

// Create runs the whole sale as one transaction.
func (r *saleRepository) Create(s *models.Sale) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Insert the invoice + items.
		if err := tx.Create(s).Error; err != nil {
			return err
		}

		// 2. Decrease stock with a GUARDED, atomic update. The
		//    "AND quantity >= ?" clause means the row only updates when there
		//    is enough stock; RowsAffected == 0 tells us there was not. This is
		//    race-safe: two concurrent sales of the last unit cannot both win.
		for _, item := range s.Items {
			res := tx.Model(&models.Product{}).
				Where("id = ? AND quantity >= ?", item.ProductID, item.Quantity).
				Update("quantity", gorm.Expr("quantity - ?", item.Quantity))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return ErrInsufficientStock // rolls everything back
			}
		}

		// 3. Add the unpaid amount to the customer's running due.
		if s.Due > 0 {
			err := tx.Model(&models.Customer{}).
				Where("id = ?", s.CustomerID).
				Update("due", gorm.Expr("due + ?", s.Due)).Error
			if err != nil {
				return err
			}
		}

		return nil // commit
	})
}

func (r *saleRepository) CountAll() (int64, error) {
	var n int64
	err := r.db.Model(&models.Sale{}).Count(&n).Error
	return n, err
}

func (r *saleRepository) FindAll(search string, offset, limit int) ([]models.Sale, int64, error) {
	var sales []models.Sale
	var total int64

	query := r.db.Model(&models.Sale{})
	if search != "" {
		query = query.Where("invoice_no LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Customer").
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&sales).Error
	if err != nil {
		return nil, 0, err
	}
	return sales, total, nil
}

func (r *saleRepository) FindByID(id uint) (*models.Sale, error) {
	var sale models.Sale
	err := r.db.
		Preload("Customer").
		Preload("Items.Product").
		First(&sale, id).Error
	if err != nil {
		return nil, err
	}
	return &sale, nil
}

// Delete VOIDS a sale: in one transaction it returns the stock it removed and
// reverses the customer due it created, then soft-deletes the invoice + items.
func (r *saleRepository) Delete(id uint) error {
	var s models.Sale
	if err := r.db.Preload("Items").First(&s, id).Error; err != nil {
		return err
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, it := range s.Items {
			err := tx.Model(&models.Product{}).Where("id = ?", it.ProductID).
				Update("quantity", gorm.Expr("quantity + ?", it.Quantity)).Error
			if err != nil {
				return err
			}
		}
		if s.Due > 0 {
			err := tx.Model(&models.Customer{}).Where("id = ?", s.CustomerID).
				Update("due", gorm.Expr("GREATEST(due - ?, 0)", s.Due)).Error
			if err != nil {
				return err
			}
		}
		if err := tx.Where("sale_id = ?", s.ID).Delete(&models.SaleItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Sale{}, s.ID).Error
	})
}
