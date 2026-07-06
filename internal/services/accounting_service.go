package services

import (
	"errors"
	"math"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

// ---- Trial Balance ----

type TrialBalanceRow struct {
	AccountID uint    `json:"account_id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Debit     float64 `json:"debit"`
	Credit    float64 `json:"credit"`
}

type TrialBalance struct {
	Rows        []TrialBalanceRow `json:"rows"`
	TotalDebit  float64           `json:"total_debit"`
	TotalCredit float64           `json:"total_credit"`
}

// ---- General Ledger ----

type GeneralLedgerEntry struct {
	Date    string  `json:"date"`
	EntryNo string  `json:"entry_no"`
	Ref     string  `json:"ref"`
	Note    string  `json:"note"`
	Debit   float64 `json:"debit"`
	Credit  float64 `json:"credit"`
	Balance float64 `json:"balance"`
}

type GeneralLedger struct {
	Account *models.Account      `json:"account"`
	Entries []GeneralLedgerEntry `json:"entries"`
	Closing float64              `json:"closing"`
}

type AccountingService interface {
	TrialBalance() (*TrialBalance, error)
	GeneralLedger(accountID uint) (*GeneralLedger, error)
}

type accountingService struct {
	journalRepo repositories.JournalRepository
	accountRepo repositories.AccountRepository
}

func NewAccountingService(journalRepo repositories.JournalRepository, accountRepo repositories.AccountRepository) AccountingService {
	return &accountingService{journalRepo: journalRepo, accountRepo: accountRepo}
}

func (s *accountingService) TrialBalance() (*TrialBalance, error) {
	balances, err := s.journalRepo.TrialBalance()
	if err != nil {
		return nil, err
	}

	// Fetch all accounts once for code/name/type lookup.
	accounts, _, err := s.accountRepo.FindAll("", "", 0, 100000)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint]models.Account, len(accounts))
	for _, a := range accounts {
		byID[a.ID] = a
	}

	out := &TrialBalance{}
	for _, b := range balances {
		acc := byID[b.AccountID]
		// Net each account into a single column (standard trial balance).
		net := b.Debit - b.Credit
		if math.Abs(net) < 0.005 {
			continue // no net balance -> omit
		}
		row := TrialBalanceRow{AccountID: b.AccountID, Code: acc.Code, Name: acc.Name, Type: acc.Type}
		if net > 0 {
			row.Debit = net
			out.TotalDebit += net
		} else {
			row.Credit = -net
			out.TotalCredit += -net
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}

func (s *accountingService) GeneralLedger(accountID uint) (*GeneralLedger, error) {
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	lines, err := s.journalRepo.AccountLedger(accountID)
	if err != nil {
		return nil, err
	}

	debitNormal := isDebitNormal(account.Type)
	out := &GeneralLedger{Account: account}
	var bal float64
	for _, l := range lines {
		if debitNormal {
			bal += l.Debit - l.Credit
		} else {
			bal += l.Credit - l.Debit
		}
		out.Entries = append(out.Entries, GeneralLedgerEntry{
			Date: l.Date.Format("2006-01-02"), EntryNo: l.EntryNo, Ref: l.Reference, Note: l.Note,
			Debit: l.Debit, Credit: l.Credit, Balance: bal,
		})
	}
	out.Closing = bal
	return out, nil
}
