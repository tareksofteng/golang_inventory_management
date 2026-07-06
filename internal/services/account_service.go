package services

import (
	"errors"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrAccountCodeTaken   = errors.New("account code already exists")
	ErrInvalidAccountType = errors.New("invalid account type")
)

// validAccountTypes and their normal (increasing) side.
var debitNormalTypes = map[string]bool{"asset": true, "expense": true}
var creditNormalTypes = map[string]bool{"liability": true, "equity": true, "income": true}

func isValidAccountType(t string) bool { return debitNormalTypes[t] || creditNormalTypes[t] }

// isDebitNormal reports whether an account of this type increases with a debit.
func isDebitNormal(t string) bool { return debitNormalTypes[t] }

type AccountService interface {
	Create(account *models.Account) (*models.Account, error)
	List(search, accType string, page, perPage int) ([]models.Account, int64, error)
	ListActive() ([]models.Account, error)
	Get(id uint) (*models.Account, error)
	Update(id uint, data *models.Account) (*models.Account, error)
	Delete(id uint) error
}

type accountService struct {
	repo repositories.AccountRepository
}

func NewAccountService(repo repositories.AccountRepository) AccountService {
	return &accountService{repo: repo}
}

func (s *accountService) Create(account *models.Account) (*models.Account, error) {
	if !isValidAccountType(account.Type) {
		return nil, ErrInvalidAccountType
	}
	exists, err := s.repo.ExistsByCode(account.Code, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAccountCodeTaken
	}
	if err := s.repo.Create(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *accountService) List(search, accType string, page, perPage int) ([]models.Account, int64, error) {
	return s.repo.FindAll(search, accType, (page-1)*perPage, perPage)
}

func (s *accountService) ListActive() ([]models.Account, error) {
	return s.repo.FindActive()
}

func (s *accountService) Get(id uint) (*models.Account, error) {
	account, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return account, nil
}

func (s *accountService) Update(id uint, data *models.Account) (*models.Account, error) {
	if !isValidAccountType(data.Type) {
		return nil, ErrInvalidAccountType
	}
	existing, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	exists, err := s.repo.ExistsByCode(data.Code, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAccountCodeTaken
	}

	existing.Code = data.Code
	existing.Name = data.Name
	existing.Type = data.Type
	existing.IsActive = data.IsActive
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *accountService) Delete(id uint) error {
	if _, err := s.Get(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
