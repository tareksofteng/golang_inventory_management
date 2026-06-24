package services

import (
	"errors"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

// Sentinel errors for the supplier domain. The unique field is EMAIL, so the
// "taken" error is about email, not name.
var (
	ErrSupplierNotFound   = errors.New("supplier not found")
	ErrSupplierEmailTaken = errors.New("supplier email already exists")
)

// SupplierService holds the business rules for suppliers.
type SupplierService interface {
	Create(supplier *models.Supplier) (*models.Supplier, error)
	List(search string, page, perPage int) ([]models.Supplier, int64, error)
	Get(id uint) (*models.Supplier, error)
	Update(id uint, data *models.Supplier) (*models.Supplier, error)
	Delete(id uint) error
}

type supplierService struct {
	repo repositories.SupplierRepository
}

// NewSupplierService injects the repository (as an interface) into the service.
func NewSupplierService(repo repositories.SupplierRepository) SupplierService {
	return &supplierService{repo: repo}
}

// Create applies the "email must be unique" business rule, then persists.
func (s *supplierService) Create(supplier *models.Supplier) (*models.Supplier, error) {
	exists, err := s.repo.ExistsByEmail(supplier.Email, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSupplierEmailTaken
	}

	if err := s.repo.Create(supplier); err != nil {
		return nil, err
	}
	return supplier, nil
}

// List converts page/perPage into a DB offset and delegates to the repo.
func (s *supplierService) List(search string, page, perPage int) ([]models.Supplier, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.FindAll(search, offset, perPage)
}

// Get returns one supplier, translating GORM's "not found" into our sentinel.
func (s *supplierService) Get(id uint) (*models.Supplier, error) {
	supplier, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, err
	}
	return supplier, nil
}

// Update loads the existing row, enforces unique-email (excluding itself), then
// saves the changed fields. We mutate the loaded record so CreatedAt and other
// untouched columns are preserved.
func (s *supplierService) Update(id uint, data *models.Supplier) (*models.Supplier, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, err
	}

	exists, err := s.repo.ExistsByEmail(data.Email, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSupplierEmailTaken
	}

	existing.Name = data.Name
	existing.Email = data.Email
	existing.Phone = data.Phone
	existing.Address = data.Address
	existing.IsActive = data.IsActive

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Delete ensures the supplier exists (for a clean 404) before soft-deleting.
func (s *supplierService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSupplierNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
