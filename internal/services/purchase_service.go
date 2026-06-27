package services

import (
	"errors"
	"fmt"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

var (
	ErrPurchaseNotFound = errors.New("purchase not found")
	ErrNoItems          = errors.New("at least one item is required")
	ErrPaidExceedsTotal = errors.New("paid amount cannot exceed total amount")
)

// PurchaseItemInput is one requested line.
type PurchaseItemInput struct {
	ProductID uint
	Quantity  int
	UnitCost  float64
}

// CreatePurchaseInput is what the controller hands the service.
type CreatePurchaseInput struct {
	SupplierID uint
	UserID     uint
	PaidAmount float64
	Note       string
	Items      []PurchaseItemInput
}

type PurchaseService interface {
	Create(input CreatePurchaseInput) (*models.Purchase, error)
	List(search string, page, perPage int) ([]models.Purchase, int64, error)
	Get(id uint) (*models.Purchase, error)
}

type purchaseService struct {
	repo         repositories.PurchaseRepository
	supplierRepo repositories.SupplierRepository
	productRepo  repositories.ProductRepository
}

func NewPurchaseService(
	repo repositories.PurchaseRepository,
	supplierRepo repositories.SupplierRepository,
	productRepo repositories.ProductRepository,
) PurchaseService {
	return &purchaseService{repo: repo, supplierRepo: supplierRepo, productRepo: productRepo}
}

func (s *purchaseService) Create(input CreatePurchaseInput) (*models.Purchase, error) {
	if len(input.Items) == 0 {
		return nil, ErrNoItems
	}

	// Referenced supplier must exist.
	if _, err := s.supplierRepo.FindByID(input.SupplierID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, err
	}

	// Build the line items, validating each product and computing totals.
	var total float64
	items := make([]models.PurchaseItem, 0, len(input.Items))
	for _, in := range input.Items {
		if in.Quantity <= 0 {
			return nil, fmt.Errorf("quantity must be greater than 0")
		}
		if _, err := s.productRepo.FindByID(in.ProductID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrProductNotFound
			}
			return nil, err
		}

		subtotal := float64(in.Quantity) * in.UnitCost
		total += subtotal
		items = append(items, models.PurchaseItem{
			ProductID: in.ProductID,
			Quantity:  in.Quantity,
			UnitCost:  in.UnitCost,
			Subtotal:  subtotal,
		})
	}

	if input.PaidAmount > total {
		return nil, ErrPaidExceedsTotal
	}

	// Sequential, human-readable invoice number (PUR-000001).
	count, err := s.repo.CountAll()
	if err != nil {
		return nil, err
	}

	purchase := &models.Purchase{
		InvoiceNo:   fmt.Sprintf("PUR-%06d", count+1),
		SupplierID:  input.SupplierID,
		UserID:      input.UserID,
		TotalAmount: total,
		PaidAmount:  input.PaidAmount,
		Due:         total - input.PaidAmount,
		Note:        input.Note,
		Items:       items,
	}

	if err := s.repo.Create(purchase); err != nil {
		return nil, err
	}
	// Re-fetch with associations for a complete response.
	return s.repo.FindByID(purchase.ID)
}

func (s *purchaseService) List(search string, page, perPage int) ([]models.Purchase, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.FindAll(search, offset, perPage)
}

func (s *purchaseService) Get(id uint) (*models.Purchase, error) {
	purchase, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPurchaseNotFound
		}
		return nil, err
	}
	return purchase, nil
}
