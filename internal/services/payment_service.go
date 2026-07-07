package services

import (
	"errors"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

var ErrAmountExceedsDue = errors.New("amount exceeds outstanding due")

type PaymentService interface {
	PayCustomer(customerID uint, amount float64, method, note string, userID uint) (*models.Payment, error)
	PaySupplier(supplierID uint, amount float64, method, note string, userID uint) (*models.Payment, error)
	List(partyType string, page, perPage int) ([]models.Payment, int64, error)
}

type paymentService struct {
	repo         repositories.PaymentRepository
	customerRepo repositories.CustomerRepository
	supplierRepo repositories.SupplierRepository
	poster       AccountingPoster
}

func NewPaymentService(
	repo repositories.PaymentRepository,
	customerRepo repositories.CustomerRepository,
	supplierRepo repositories.SupplierRepository,
	poster AccountingPoster,
) PaymentService {
	return &paymentService{repo: repo, customerRepo: customerRepo, supplierRepo: supplierRepo, poster: poster}
}

func (s *paymentService) PayCustomer(customerID uint, amount float64, method, note string, userID uint) (*models.Payment, error) {
	customer, err := s.customerRepo.FindByID(customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	payment := &models.Payment{
		PartyType: "customer",
		PartyID:   customer.ID,
		PartyName: customer.Name,
		UserID:    userID,
		Amount:    amount,
		Method:    method,
		Note:      note,
	}
	if err := s.repo.CreateCustomerPayment(payment); err != nil {
		if errors.Is(err, repositories.ErrAmountExceedsDue) {
			return nil, ErrAmountExceedsDue
		}
		return nil, err
	}
	if s.poster != nil {
		s.poster.PostCustomerPayment(payment)
	}
	return payment, nil
}

func (s *paymentService) PaySupplier(supplierID uint, amount float64, method, note string, userID uint) (*models.Payment, error) {
	supplier, err := s.supplierRepo.FindByID(supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, err
	}

	payment := &models.Payment{
		PartyType: "supplier",
		PartyID:   supplier.ID,
		PartyName: supplier.Name,
		UserID:    userID,
		Amount:    amount,
		Method:    method,
		Note:      note,
	}
	if err := s.repo.CreateSupplierPayment(payment); err != nil {
		if errors.Is(err, repositories.ErrAmountExceedsDue) {
			return nil, ErrAmountExceedsDue
		}
		return nil, err
	}
	if s.poster != nil {
		s.poster.PostSupplierPayment(payment)
	}
	return payment, nil
}

func (s *paymentService) List(partyType string, page, perPage int) ([]models.Payment, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.FindAll(partyType, offset, perPage)
}
