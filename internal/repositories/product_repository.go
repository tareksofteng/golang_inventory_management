package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// ProductRepository describes data-access for products. Reads preload the
// Category and Supplier associations so the API can return nested objects.
type ProductRepository interface {
	Create(product *models.Product) error
	FindAll(search string, offset, limit int) ([]models.Product, int64, error)
	FindByID(id uint) (*models.Product, error)
	Update(product *models.Product) error
	Delete(id uint) error
	ExistsBySKU(sku string, excludeID uint) (bool, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

// FindAll returns paginated products (searched by name or SKU) with their
// Category and Supplier eager-loaded. Preload runs a second query and stitches
// the results — the Go way to avoid N+1, chosen explicitly per call.
func (r *productRepository) FindAll(search string, offset, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR sku LIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Category").
		Preload("Supplier").
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

// FindByID returns one product with Category + Supplier preloaded.
func (r *productRepository) FindByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.
		Preload("Category").
		Preload("Supplier").
		First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

// ExistsBySKU reports whether another product already uses this SKU.
func (r *productRepository) ExistsBySKU(sku string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Product{}).Where("sku = ?", sku)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
