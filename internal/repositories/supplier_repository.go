package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// SupplierRepository describes the data-access operations for suppliers.
// We program against this INTERFACE (not the concrete struct) so the service
// layer can be unit-tested with a fake repo, and so the DB technology stays
// swappable. Think of it as a Laravel "contract".
type SupplierRepository interface {
	Create(supplier *models.Supplier) error
	FindAll(search string, offset, limit int) ([]models.Supplier, int64, error)
	FindByID(id uint) (*models.Supplier, error)
	Update(supplier *models.Supplier) error
	Delete(id uint) error
	ExistsByEmail(email string, excludeID uint) (bool, error)
}

// supplierRepository is the GORM-backed implementation. lowercase = private:
// callers receive it only through the interface, never by its concrete type.
type supplierRepository struct {
	db *gorm.DB
}

// NewSupplierRepository wires a *gorm.DB into a SupplierRepository.
func NewSupplierRepository(db *gorm.DB) SupplierRepository {
	return &supplierRepository{db: db}
}

// Create inserts a new supplier. GORM fills ID/CreatedAt/UpdatedAt on success.
func (r *supplierRepository) Create(supplier *models.Supplier) error {
	return r.db.Create(supplier).Error
}

// FindAll returns a paginated, optionally searched slice of suppliers plus the
// TOTAL matching count. Search matches either the name OR the email.
func (r *supplierRepository) FindAll(search string, offset, limit int) ([]models.Supplier, int64, error) {
	var suppliers []models.Supplier
	var total int64

	query := r.db.Model(&models.Supplier{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR email LIKE ?", like, like)
	}

	// Count BEFORE applying offset/limit, otherwise total would be wrong.
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&suppliers).Error
	if err != nil {
		return nil, 0, err
	}
	return suppliers, total, nil
}

// FindByID returns a single supplier or gorm.ErrRecordNotFound when missing.
func (r *supplierRepository) FindByID(id uint) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := r.db.First(&supplier, id).Error; err != nil {
		return nil, err
	}
	return &supplier, nil
}

// Update persists changes to an existing supplier. Save writes all fields.
func (r *supplierRepository) Update(supplier *models.Supplier) error {
	return r.db.Save(supplier).Error
}

// Delete soft-deletes by primary key (sets deleted_at, keeps the row).
func (r *supplierRepository) Delete(id uint) error {
	return r.db.Delete(&models.Supplier{}, id).Error
}

// ExistsByEmail reports whether another supplier already uses this email.
// An empty email is never treated as a duplicate (multiple suppliers may have
// no email). excludeID lets an update skip its own row (0 = check all rows).
func (r *supplierRepository) ExistsByEmail(email string, excludeID uint) (bool, error) {
	if email == "" {
		return false, nil
	}

	var count int64
	query := r.db.Model(&models.Supplier{}).Where("email = ?", email)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
