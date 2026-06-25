package services

import (
	"errors"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

var ErrCustomerNotFound = errors.New("customer not found")

type CustomerService interface {
	Create(customer *models.Customer) (*models.Customer, error)
	List(search string, page, perPage int) ([]models.Customer, int64, error)
	Get(id uint) (*models.Customer, error)
	Update(id uint, data *models.Customer) (*models.Customer, error)
	Delete(id uint) error
}

type customerService struct {
	repo repositories.CustomerRepository
}

func NewCustomerService(repo repositories.CustomerRepository) CustomerService {
	return &customerService{repo: repo}
}

func (s *customerService) Create(customer *models.Customer) (*models.Customer, error) {
	if err := s.repo.Create(customer); err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *customerService) List(search string, page, perPage int) ([]models.Customer, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.FindAll(search, offset, perPage)
}

func (s *customerService) Get(id uint) (*models.Customer, error) {
	customer, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return customer, nil
}

// Update changes profile fields. Due is editable here (manual adjustment of the
// opening balance); once Sales exists, sales/payments will move it instead.
func (s *customerService) Update(id uint, data *models.Customer) (*models.Customer, error) {
	existing, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	existing.Name = data.Name
	existing.Email = data.Email
	existing.Phone = data.Phone
	existing.Address = data.Address
	existing.Due = data.Due
	existing.IsActive = data.IsActive

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *customerService) Delete(id uint) error {
	if _, err := s.Get(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
