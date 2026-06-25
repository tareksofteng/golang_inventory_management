package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

type CustomerRepository interface {
	Create(customer *models.Customer) error
	FindAll(search string, offset, limit int) ([]models.Customer, int64, error)
	FindByID(id uint) (*models.Customer, error)
	Update(customer *models.Customer) error
	Delete(id uint) error
}

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(customer *models.Customer) error {
	return r.db.Create(customer).Error
}

func (r *customerRepository) FindAll(search string, offset, limit int) ([]models.Customer, int64, error) {
	var customers []models.Customer
	var total int64

	query := r.db.Model(&models.Customer{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR email LIKE ? OR phone LIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&customers).Error
	if err != nil {
		return nil, 0, err
	}
	return customers, total, nil
}

func (r *customerRepository) FindByID(id uint) (*models.Customer, error) {
	var customer models.Customer
	if err := r.db.First(&customer, id).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) Update(customer *models.Customer) error {
	return r.db.Save(customer).Error
}

func (r *customerRepository) Delete(id uint) error {
	return r.db.Delete(&models.Customer{}, id).Error
}
