package services

import (
	"errors"
	"fmt"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

var (
	ErrSaleNotFound      = errors.New("sale not found")
	ErrInsufficientStock = errors.New("insufficient stock for one or more items")
)

type SaleItemInput struct {
	ProductID uint
	Quantity  int
	UnitPrice float64
}

type CreateSaleInput struct {
	CustomerID uint
	UserID     uint
	PaidAmount float64
	Note       string
	Items      []SaleItemInput
}

type SaleService interface {
	Create(input CreateSaleInput) (*models.Sale, error)
	List(search string, page, perPage int) ([]models.Sale, int64, error)
	Get(id uint) (*models.Sale, error)
}

type saleService struct {
	repo         repositories.SaleRepository
	customerRepo repositories.CustomerRepository
	productRepo  repositories.ProductRepository
}

func NewSaleService(
	repo repositories.SaleRepository,
	customerRepo repositories.CustomerRepository,
	productRepo repositories.ProductRepository,
) SaleService {
	return &saleService{repo: repo, customerRepo: customerRepo, productRepo: productRepo}
}

func (s *saleService) Create(input CreateSaleInput) (*models.Sale, error) {
	if len(input.Items) == 0 {
		return nil, ErrNoItems
	}

	if _, err := s.customerRepo.FindByID(input.CustomerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	// Build items, validating products and computing totals. We also do a
	// friendly pre-check of stock here so the common case returns a clear
	// message; the repository's transaction is still the authoritative,
	// race-safe guard.
	var total float64
	items := make([]models.SaleItem, 0, len(input.Items))
	for _, in := range input.Items {
		if in.Quantity <= 0 {
			return nil, fmt.Errorf("quantity must be greater than 0")
		}
		product, err := s.productRepo.FindByID(in.ProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrProductNotFound
			}
			return nil, err
		}
		if product.Quantity < in.Quantity {
			return nil, ErrInsufficientStock
		}

		subtotal := float64(in.Quantity) * in.UnitPrice
		total += subtotal
		items = append(items, models.SaleItem{
			ProductID: in.ProductID,
			Quantity:  in.Quantity,
			UnitPrice: in.UnitPrice,
			Subtotal:  subtotal,
		})
	}

	if input.PaidAmount > total {
		return nil, ErrPaidExceedsTotal
	}

	count, err := s.repo.CountAll()
	if err != nil {
		return nil, err
	}

	sale := &models.Sale{
		InvoiceNo:   fmt.Sprintf("SAL-%06d", count+1),
		CustomerID:  input.CustomerID,
		UserID:      input.UserID,
		TotalAmount: total,
		PaidAmount:  input.PaidAmount,
		Due:         total - input.PaidAmount,
		Note:        input.Note,
		Items:       items,
	}

	if err := s.repo.Create(sale); err != nil {
		// Translate the repository's race-safe stock error to ours.
		if errors.Is(err, repositories.ErrInsufficientStock) {
			return nil, ErrInsufficientStock
		}
		return nil, err
	}
	return s.repo.FindByID(sale.ID)
}

func (s *saleService) List(search string, page, perPage int) ([]models.Sale, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.FindAll(search, offset, perPage)
}

func (s *saleService) Get(id uint) (*models.Sale, error) {
	sale, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSaleNotFound
		}
		return nil, err
	}
	return sale, nil
}
