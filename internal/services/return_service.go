package services

import (
	"errors"
	"fmt"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

// ReturnItemInput is one return line. UnitValue is the cost (purchase return) or
// price (sale return) per unit — the caller decides the meaning.
type ReturnItemInput struct {
	ProductID uint
	Quantity  int
	UnitValue float64
}

type CreatePurchaseReturnInput struct {
	SupplierID uint
	UserID     uint
	Note       string
	Items      []ReturnItemInput
}

type CreateSaleReturnInput struct {
	CustomerID uint
	UserID     uint
	Note       string
	Items      []ReturnItemInput
}

type ReturnService interface {
	CreatePurchaseReturn(input CreatePurchaseReturnInput) (*models.PurchaseReturn, error)
	CreateSaleReturn(input CreateSaleReturnInput) (*models.SaleReturn, error)
	ListPurchaseReturns(page, perPage int) ([]models.PurchaseReturn, int64, error)
	ListSaleReturns(page, perPage int) ([]models.SaleReturn, int64, error)
}

type returnService struct {
	repo         repositories.ReturnRepository
	supplierRepo repositories.SupplierRepository
	customerRepo repositories.CustomerRepository
	productRepo  repositories.ProductRepository
}

func NewReturnService(
	repo repositories.ReturnRepository,
	supplierRepo repositories.SupplierRepository,
	customerRepo repositories.CustomerRepository,
	productRepo repositories.ProductRepository,
) ReturnService {
	return &returnService{repo: repo, supplierRepo: supplierRepo, customerRepo: customerRepo, productRepo: productRepo}
}

// validateItems checks each line's product exists and quantity is positive,
// returning the validated lines + grand total.
func (s *returnService) validateItems(items []ReturnItemInput) ([]ReturnItemInput, float64, error) {
	if len(items) == 0 {
		return nil, 0, ErrNoItems
	}
	var total float64
	for _, in := range items {
		if in.Quantity <= 0 {
			return nil, 0, fmt.Errorf("quantity must be greater than 0")
		}
		if _, err := s.productRepo.FindByID(in.ProductID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, ErrProductNotFound
			}
			return nil, 0, err
		}
		total += float64(in.Quantity) * in.UnitValue
	}
	return items, total, nil
}

func (s *returnService) CreatePurchaseReturn(input CreatePurchaseReturnInput) (*models.PurchaseReturn, error) {
	if _, err := s.supplierRepo.FindByID(input.SupplierID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, err
	}

	lines, total, err := s.validateItems(input.Items)
	if err != nil {
		return nil, err
	}

	items := make([]models.PurchaseReturnItem, len(lines))
	for i, in := range lines {
		items[i] = models.PurchaseReturnItem{
			ProductID: in.ProductID,
			Quantity:  in.Quantity,
			UnitCost:  in.UnitValue,
			Subtotal:  float64(in.Quantity) * in.UnitValue,
		}
	}

	count, err := s.repo.CountPurchaseReturns()
	if err != nil {
		return nil, err
	}

	ret := &models.PurchaseReturn{
		InvoiceNo:   fmt.Sprintf("PRET-%06d", count+1),
		SupplierID:  input.SupplierID,
		UserID:      input.UserID,
		TotalAmount: total,
		Note:        input.Note,
		Items:       items,
	}
	if err := s.repo.CreatePurchaseReturn(ret); err != nil {
		if errors.Is(err, repositories.ErrInsufficientStock) {
			return nil, ErrInsufficientStock
		}
		return nil, err
	}
	return ret, nil
}

func (s *returnService) CreateSaleReturn(input CreateSaleReturnInput) (*models.SaleReturn, error) {
	if _, err := s.customerRepo.FindByID(input.CustomerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	lines, total, err := s.validateItems(input.Items)
	if err != nil {
		return nil, err
	}

	items := make([]models.SaleReturnItem, len(lines))
	for i, in := range lines {
		items[i] = models.SaleReturnItem{
			ProductID: in.ProductID,
			Quantity:  in.Quantity,
			UnitPrice: in.UnitValue,
			Subtotal:  float64(in.Quantity) * in.UnitValue,
		}
	}

	count, err := s.repo.CountSaleReturns()
	if err != nil {
		return nil, err
	}

	ret := &models.SaleReturn{
		InvoiceNo:   fmt.Sprintf("SRET-%06d", count+1),
		CustomerID:  input.CustomerID,
		UserID:      input.UserID,
		TotalAmount: total,
		Note:        input.Note,
		Items:       items,
	}
	if err := s.repo.CreateSaleReturn(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *returnService) ListPurchaseReturns(page, perPage int) ([]models.PurchaseReturn, int64, error) {
	return s.repo.FindPurchaseReturns((page-1)*perPage, perPage)
}

func (s *returnService) ListSaleReturns(page, perPage int) ([]models.SaleReturn, int64, error) {
	return s.repo.FindSaleReturns((page-1)*perPage, perPage)
}
