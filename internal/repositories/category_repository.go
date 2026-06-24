package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// CategoryRepository describes the data-access operations for categories.
// We program against this INTERFACE (not the concrete struct) so the service
// layer can be unit-tested with a fake repo, and so the DB technology stays
// swappable. Think of it as a Laravel "contract".
type CategoryRepository interface {
	Create(category *models.Category) error
	FindAll(search string, offset, limit int) ([]models.Category, int64, error)
	FindByID(id uint) (*models.Category, error)
	Update(category *models.Category) error
	Delete(id uint) error
	ExistsByName(name string, excludeID uint) (bool, error)
}

// categoryRepository is the GORM-backed implementation. It is lowercase
// (unexported) on purpose — callers receive it only through the interface,
// never by its concrete type.
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository wires a *gorm.DB into a CategoryRepository. This is
// constructor-style dependency injection — main() builds it and passes it down.
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// Create inserts a new category. GORM fills ID/CreatedAt/UpdatedAt on success.
func (r *categoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

// FindAll returns a paginated, optionally name-searched slice of categories
// plus the TOTAL matching count (needed to compute total pages).
//
// Soft delete is automatic: GORM adds `WHERE deleted_at IS NULL` for us.
func (r *categoryRepository) FindAll(search string, offset, limit int) ([]models.Category, int64, error) {
	var categories []models.Category
	var total int64

	query := r.db.Model(&models.Category{})
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	// Count BEFORE applying offset/limit, otherwise total would be wrong.
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}
	return categories, total, nil
}

// FindByID returns a single category or gorm.ErrRecordNotFound when missing.
func (r *categoryRepository) FindByID(id uint) (*models.Category, error) {
	var category models.Category
	if err := r.db.First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

// Update persists changes to an existing category. Save writes all fields.
func (r *categoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

// Delete soft-deletes by primary key (sets deleted_at, keeps the row).
func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}

// ExistsByName reports whether another category already uses this name.
// excludeID lets an update skip its own row (0 = check all rows, used on create).
func (r *categoryRepository) ExistsByName(name string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Category{}).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
