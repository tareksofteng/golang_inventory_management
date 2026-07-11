package repositories

import (
	"time"

	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// AccountBalance is the debit/credit total for one account (trial balance row).
type AccountBalance struct {
	AccountID uint
	Debit     float64
	Credit    float64
}

// LedgerLine is one journal posting against an account (general-ledger row).
type LedgerLine struct {
	Date      time.Time
	EntryNo   string
	Reference string
	Note      string
	Debit     float64
	Credit    float64
}

type JournalRepository interface {
	Create(entry *models.JournalEntry) error
	CountAll() (int64, error)
	FindAll(offset, limit int) ([]models.JournalEntry, int64, error)
	FindByID(id uint) (*models.JournalEntry, error)
	TrialBalance() ([]AccountBalance, error)
	TrialBalanceBetween(from, to time.Time) ([]AccountBalance, error)
	TrialBalanceAsOf(to time.Time) ([]AccountBalance, error)
	AccountLedger(accountID uint) ([]LedgerLine, error)
}

type journalRepository struct {
	db *gorm.DB
}

func NewJournalRepository(db *gorm.DB) JournalRepository {
	return &journalRepository{db: db}
}

// Create inserts the entry + its lines in one transaction.
func (r *journalRepository) Create(entry *models.JournalEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(entry).Error
	})
}

func (r *journalRepository) CountAll() (int64, error) {
	var n int64
	err := r.db.Model(&models.JournalEntry{}).Count(&n).Error
	return n, err
}

func (r *journalRepository) FindAll(offset, limit int) ([]models.JournalEntry, int64, error) {
	var entries []models.JournalEntry
	var total int64
	if err := r.db.Model(&models.JournalEntry{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Preload("Lines.Account").Order("id DESC").Offset(offset).Limit(limit).Find(&entries).Error
	return entries, total, err
}

func (r *journalRepository) FindByID(id uint) (*models.JournalEntry, error) {
	var entry models.JournalEntry
	err := r.db.Preload("Lines.Account").First(&entry, id).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// TrialBalance sums debits and credits per account across all journal lines.
func (r *journalRepository) TrialBalance() ([]AccountBalance, error) {
	var rows []AccountBalance
	err := r.db.Table("journal_lines").
		Select("account_id, SUM(debit) as debit, SUM(credit) as credit").
		Where("deleted_at IS NULL").
		Group("account_id").
		Scan(&rows).Error
	return rows, err
}

// TrialBalanceBetween sums debits/credits per account for entries dated in
// [from, to) — used for period reports like Profit & Loss.
func (r *journalRepository) TrialBalanceBetween(from, to time.Time) ([]AccountBalance, error) {
	var rows []AccountBalance
	err := r.db.Table("journal_lines").
		Select("journal_lines.account_id as account_id, SUM(journal_lines.debit) as debit, SUM(journal_lines.credit) as credit").
		Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_lines.deleted_at IS NULL AND journal_entries.deleted_at IS NULL AND journal_entries.date >= ? AND journal_entries.date < ?", from, to).
		Group("journal_lines.account_id").
		Scan(&rows).Error
	return rows, err
}

// TrialBalanceAsOf sums debits/credits per account for entries dated before to
// (exclusive) — a point-in-time snapshot used by the Balance Sheet.
func (r *journalRepository) TrialBalanceAsOf(to time.Time) ([]AccountBalance, error) {
	var rows []AccountBalance
	err := r.db.Table("journal_lines").
		Select("journal_lines.account_id as account_id, SUM(journal_lines.debit) as debit, SUM(journal_lines.credit) as credit").
		Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_lines.deleted_at IS NULL AND journal_entries.deleted_at IS NULL AND journal_entries.date < ?", to).
		Group("journal_lines.account_id").
		Scan(&rows).Error
	return rows, err
}

// AccountLedger returns every posting against one account, oldest first.
func (r *journalRepository) AccountLedger(accountID uint) ([]LedgerLine, error) {
	var rows []LedgerLine
	err := r.db.Table("journal_lines").
		Select("journal_entries.date as date, journal_entries.entry_no as entry_no, journal_entries.reference as reference, journal_entries.note as note, journal_lines.debit as debit, journal_lines.credit as credit").
		Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_lines.account_id = ? AND journal_lines.deleted_at IS NULL AND journal_entries.deleted_at IS NULL", accountID).
		Order("journal_entries.date ASC, journal_entries.id ASC").
		Scan(&rows).Error
	return rows, err
}
