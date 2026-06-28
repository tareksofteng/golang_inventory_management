package services

import (
	"errors"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

// Sentinel errors the service returns so the controller can map them to the
// right HTTP status WITHOUT knowing about GORM. The controller checks these
// with errors.Is(err, ErrCategoryNotFound). This keeps HTTP concerns out of
// the service and DB concerns out of the controller.
var (
	ErrCategoryNotFound  = errors.New("category not found")
	ErrCategoryNameTaken = errors.New("category name already exists")
)

// CategoryService holds the business rules for categories.
type CategoryService interface {
	Create(category *models.Category) (*models.Category, error)
	List(search string, page, perPage int) ([]models.Category, int64, error)
	Get(id uint) (*models.Category, error)
	Update(id uint, data *models.Category) (*models.Category, error)
	Delete(id uint) error
}

type categoryService struct {
	repo repositories.CategoryRepository
}

// NewCategoryService injects the repository (as an interface) into the service.
func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

// Create applies the "name must be unique" business rule, then persists.
func (s *categoryService) Create(category *models.Category) (*models.Category, error) {
	exists, err := s.repo.ExistsByName(category.Name, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCategoryNameTaken
	}

	if err := s.repo.Create(category); err != nil {
		return nil, err
	}
	return category, nil
}

// List converts page/perPage into a DB offset and delegates to the repo.
// Normalising bad input (page < 1 etc.) happens in the controller helper.
func (s *categoryService) List(search string, page, perPage int) ([]models.Category, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.FindAll(search, offset, perPage)
}

// Get returns one category, translating GORM's "not found" into our sentinel.
func (s *categoryService) Get(id uint) (*models.Category, error) {
	category, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	return category, nil
}

// Update loads the existing row, enforces unique-name (excluding itself), then
// saves the changed fields. We mutate the loaded record so CreatedAt and other
// untouched columns are preserved.
func (s *categoryService) Update(id uint, data *models.Category) (*models.Category, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	exists, err := s.repo.ExistsByName(data.Name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCategoryNameTaken
	}

	existing.Name = data.Name
	existing.Description = data.Description
	existing.IsActive = data.IsActive

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Delete ensures the category exists (for a clean 404) before soft-deleting.
func (s *categoryService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
