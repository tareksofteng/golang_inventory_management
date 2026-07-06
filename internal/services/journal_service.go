package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

var (
	ErrJournalNotFound = errors.New("journal entry not found")
	ErrUnbalanced      = errors.New("total debit must equal total credit")
	ErrTooFewLines     = errors.New("a journal entry needs at least two lines")
	ErrInvalidLine     = errors.New("each line must have either a debit or a credit (not both, not zero)")
)

type JournalLineInput struct {
	AccountID uint
	Debit     float64
	Credit    float64
}

type CreateJournalInput struct {
	Date      time.Time
	Reference string
	Note      string
	UserID    uint
	Lines     []JournalLineInput
}

type JournalService interface {
	Create(input CreateJournalInput) (*models.JournalEntry, error)
	List(page, perPage int) ([]models.JournalEntry, int64, error)
	Get(id uint) (*models.JournalEntry, error)
}

type journalService struct {
	repo        repositories.JournalRepository
	accountRepo repositories.AccountRepository
}

func NewJournalService(repo repositories.JournalRepository, accountRepo repositories.AccountRepository) JournalService {
	return &journalService{repo: repo, accountRepo: accountRepo}
}

func (s *journalService) Create(input CreateJournalInput) (*models.JournalEntry, error) {
	if len(input.Lines) < 2 {
		return nil, ErrTooFewLines
	}

	var totalDebit, totalCredit float64
	lines := make([]models.JournalLine, 0, len(input.Lines))
	for _, in := range input.Lines {
		// Exactly one of debit/credit must be positive.
		if (in.Debit > 0) == (in.Credit > 0) {
			return nil, ErrInvalidLine
		}
		if in.Debit < 0 || in.Credit < 0 {
			return nil, ErrInvalidLine
		}
		if _, err := s.accountRepo.FindByID(in.AccountID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrAccountNotFound
			}
			return nil, err
		}
		totalDebit += in.Debit
		totalCredit += in.Credit
		lines = append(lines, models.JournalLine{AccountID: in.AccountID, Debit: in.Debit, Credit: in.Credit})
	}

	// Balanced within a cent.
	if math.Abs(totalDebit-totalCredit) > 0.005 {
		return nil, ErrUnbalanced
	}

	count, err := s.repo.CountAll()
	if err != nil {
		return nil, err
	}
	date := input.Date
	if date.IsZero() {
		date = time.Now()
	}

	entry := &models.JournalEntry{
		EntryNo:   fmt.Sprintf("JE-%06d", count+1),
		Date:      date,
		Reference: input.Reference,
		Note:      input.Note,
		UserID:    input.UserID,
		Lines:     lines,
	}
	if err := s.repo.Create(entry); err != nil {
		return nil, err
	}
	return s.repo.FindByID(entry.ID)
}

func (s *journalService) List(page, perPage int) ([]models.JournalEntry, int64, error) {
	return s.repo.FindAll((page-1)*perPage, perPage)
}

func (s *journalService) Get(id uint) (*models.JournalEntry, error) {
	entry, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrJournalNotFound
		}
		return nil, err
	}
	return entry, nil
}
