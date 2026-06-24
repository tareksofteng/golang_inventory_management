package services

import (
	"errors"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrProductSKUTaken = errors.New("product SKU already exists")
	// Reused for foreign-key validation: the referenced category/supplier
	// does not exist. (ErrCategoryNotFound / ErrSupplierNotFound are declared
	// in the category/supplier services — same package, so we reuse them.)
)

type ProductService interface {
	Create(product *models.Product) (*models.Product, error)
	List(search string, page, perPage int) ([]models.Product, int64, error)
	Get(id uint) (*models.Product, error)
	Update(id uint, data *models.Product) (*models.Product, error)
	Delete(id uint) error
}

// productService also depends on the category & supplier repositories so it can
// verify that a product's CategoryID / SupplierID actually point at real rows
// BEFORE inserting — otherwise the DB would reject it with a raw FK error, or
// (worse, since we don't add hard FK constraints) accept an orphan reference.
type productService struct {
	repo         repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
	supplierRepo repositories.SupplierRepository
}

func NewProductService(
	repo repositories.ProductRepository,
	categoryRepo repositories.CategoryRepository,
	supplierRepo repositories.SupplierRepository,
) ProductService {
	return &productService{
		repo:         repo,
		categoryRepo: categoryRepo,
		supplierRepo: supplierRepo,
	}
}

// validateRefs ensures the referenced category and supplier exist.
func (s *productService) validateRefs(categoryID, supplierID uint) error {
	if _, err := s.categoryRepo.FindByID(categoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	if _, err := s.supplierRepo.FindByID(supplierID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSupplierNotFound
		}
		return err
	}
	return nil
}

func (s *productService) Create(product *models.Product) (*models.Product, error) {
	if err := s.validateRefs(product.CategoryID, product.SupplierID); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsBySKU(product.SKU, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductSKUTaken
	}

	if err := s.repo.Create(product); err != nil {
		return nil, err
	}
	// Re-fetch so the response includes the preloaded Category + Supplier.
	return s.repo.FindByID(product.ID)
}

func (s *productService) List(search string, page, perPage int) ([]models.Product, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.FindAll(search, offset, perPage)
}

func (s *productService) Get(id uint) (*models.Product, error) {
	product, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return product, nil
}

// Update changes product details but NOT Quantity — stock levels only move
// through the Stock In module, never via a plain product edit. We preserve
// existing.Quantity by simply not overwriting it.
func (s *productService) Update(id uint, data *models.Product) (*models.Product, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	if err := s.validateRefs(data.CategoryID, data.SupplierID); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsBySKU(data.SKU, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductSKUTaken
	}

	existing.Name = data.Name
	existing.SKU = data.SKU
	existing.CategoryID = data.CategoryID
	existing.SupplierID = data.SupplierID
	existing.Price = data.Price
	existing.CostPrice = data.CostPrice
	existing.Unit = data.Unit
	existing.IsActive = data.IsActive
	// Quantity intentionally NOT updated here.

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return s.repo.FindByID(existing.ID)
}

func (s *productService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
