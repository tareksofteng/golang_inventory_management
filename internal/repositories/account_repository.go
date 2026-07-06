package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

type AccountRepository interface {
	Create(account *models.Account) error
	FindAll(search, accType string, offset, limit int) ([]models.Account, int64, error)
	FindActive() ([]models.Account, error)
	FindByID(id uint) (*models.Account, error)
	Update(account *models.Account) error
	Delete(id uint) error
	ExistsByCode(code string, excludeID uint) (bool, error)
	Count() (int64, error)
}

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

func (r *accountRepository) FindAll(search, accType string, offset, limit int) ([]models.Account, int64, error) {
	var accounts []models.Account
	var total int64

	query := r.db.Model(&models.Account{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	if accType != "" {
		query = query.Where("type = ?", accType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("code ASC").Offset(offset).Limit(limit).Find(&accounts).Error
	return accounts, total, err
}

// FindActive lists all active accounts (for journal-entry dropdowns).
func (r *accountRepository) FindActive() ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Where("is_active = ?", true).Order("code ASC").Find(&accounts).Error
	return accounts, err
}

func (r *accountRepository) FindByID(id uint) (*models.Account, error) {
	var account models.Account
	if err := r.db.First(&account, id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepository) Update(account *models.Account) error {
	return r.db.Save(account).Error
}

func (r *accountRepository) Delete(id uint) error {
	return r.db.Delete(&models.Account{}, id).Error
}

func (r *accountRepository) ExistsByCode(code string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Account{}).Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *accountRepository) Count() (int64, error) {
	var n int64
	err := r.db.Model(&models.Account{}).Count(&n).Error
	return n, err
}
