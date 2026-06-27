package repositories

import (
	"errors"

	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// ErrAmountExceedsDue is returned when a payment is larger than what is owed.
var ErrAmountExceedsDue = errors.New("amount exceeds outstanding due")

type PaymentRepository interface {
	CreateCustomerPayment(p *models.Payment) error
	CreateSupplierPayment(p *models.Payment) error
	FindAll(partyType string, offset, limit int) ([]models.Payment, int64, error)
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// CreateCustomerPayment records a receipt and reduces the customer's due, in one
// transaction. The guarded update (WHERE due >= amount) prevents reducing a due
// below zero and is race-safe.
func (r *paymentRepository) CreateCustomerPayment(p *models.Payment) error {
	return r.createPayment(p, &models.Customer{})
}

// CreateSupplierPayment records a payment and reduces the supplier's due.
func (r *paymentRepository) CreateSupplierPayment(p *models.Payment) error {
	return r.createPayment(p, &models.Supplier{})
}

// createPayment is the shared transactional body for both sides.
func (r *paymentRepository) createPayment(p *models.Payment, party interface{}) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}

		res := tx.Model(party).
			Where("id = ? AND due >= ?", p.PartyID, p.Amount).
			Update("due", gorm.Expr("due - ?", p.Amount))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrAmountExceedsDue // rolls back the payment insert
		}
		return nil
	})
}

func (r *paymentRepository) FindAll(partyType string, offset, limit int) ([]models.Payment, int64, error) {
	var payments []models.Payment
	var total int64

	query := r.db.Model(&models.Payment{})
	if partyType != "" {
		query = query.Where("party_type = ?", partyType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&payments).Error
	if err != nil {
		return nil, 0, err
	}
	return payments, total, nil
}
