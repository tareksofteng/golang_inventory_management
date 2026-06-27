package services

import (
	"errors"
	"sort"
	"time"

	"inventory-api/internal/repositories"

	"gorm.io/gorm"
)

// LedgerEntry is one line of a running-balance statement.
type LedgerEntry struct {
	Date    string  `json:"date"`
	Type    string  `json:"type"`
	Ref     string  `json:"ref"`
	Debit   float64 `json:"debit"`
	Credit  float64 `json:"credit"`
	Balance float64 `json:"balance"`
}

// Ledger is a party's full statement with closing balance.
type Ledger struct {
	PartyID        uint          `json:"party_id"`
	PartyName      string        `json:"party_name"`
	Entries        []LedgerEntry `json:"entries"`
	ClosingBalance float64       `json:"closing_balance"`
}

// rawEntry carries the sortable timestamp before formatting.
type rawEntry struct {
	t      time.Time
	Type   string
	Ref    string
	Debit  float64
	Credit float64
}

type LedgerService interface {
	CustomerLedger(customerID uint) (*Ledger, error)
	SupplierLedger(supplierID uint) (*Ledger, error)
}

type ledgerService struct {
	repo         repositories.LedgerRepository
	customerRepo repositories.CustomerRepository
	supplierRepo repositories.SupplierRepository
}

func NewLedgerService(
	repo repositories.LedgerRepository,
	customerRepo repositories.CustomerRepository,
	supplierRepo repositories.SupplierRepository,
) LedgerService {
	return &ledgerService{repo: repo, customerRepo: customerRepo, supplierRepo: supplierRepo}
}

// build sorts raw entries by time and computes the running balance. Debit grows
// the balance (party owes more), Credit shrinks it.
func build(partyID uint, name string, raws []rawEntry) *Ledger {
	sort.SliceStable(raws, func(i, j int) bool { return raws[i].t.Before(raws[j].t) })

	entries := make([]LedgerEntry, 0, len(raws))
	var balance float64
	for _, r := range raws {
		balance += r.Debit - r.Credit
		entries = append(entries, LedgerEntry{
			Date:    r.t.Format("2006-01-02"),
			Type:    r.Type,
			Ref:     r.Ref,
			Debit:   r.Debit,
			Credit:  r.Credit,
			Balance: balance,
		})
	}
	return &Ledger{PartyID: partyID, PartyName: name, Entries: entries, ClosingBalance: balance}
}

func (s *ledgerService) CustomerLedger(customerID uint) (*Ledger, error) {
	customer, err := s.customerRepo.FindByID(customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	sales, err := s.repo.SalesByCustomer(customerID)
	if err != nil {
		return nil, err
	}
	payments, err := s.repo.PaymentsByParty("customer", customerID)
	if err != nil {
		return nil, err
	}
	returns, err := s.repo.SaleReturnsByCustomer(customerID)
	if err != nil {
		return nil, err
	}

	var raws []rawEntry
	for _, sale := range sales {
		raws = append(raws, rawEntry{t: sale.CreatedAt, Type: "Sale", Ref: sale.InvoiceNo, Debit: sale.TotalAmount})
		if sale.PaidAmount > 0 {
			raws = append(raws, rawEntry{t: sale.CreatedAt, Type: "Receipt (at sale)", Ref: sale.InvoiceNo, Credit: sale.PaidAmount})
		}
	}
	for _, p := range payments {
		raws = append(raws, rawEntry{t: p.CreatedAt, Type: "Payment", Ref: p.Method, Credit: p.Amount})
	}
	for _, r := range returns {
		raws = append(raws, rawEntry{t: r.CreatedAt, Type: "Sale Return", Ref: r.InvoiceNo, Credit: r.TotalAmount})
	}

	return build(customer.ID, customer.Name, raws), nil
}

func (s *ledgerService) SupplierLedger(supplierID uint) (*Ledger, error) {
	supplier, err := s.supplierRepo.FindByID(supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, err
	}

	purchases, err := s.repo.PurchasesBySupplier(supplierID)
	if err != nil {
		return nil, err
	}
	payments, err := s.repo.PaymentsByParty("supplier", supplierID)
	if err != nil {
		return nil, err
	}
	returns, err := s.repo.PurchaseReturnsBySupplier(supplierID)
	if err != nil {
		return nil, err
	}

	var raws []rawEntry
	for _, p := range purchases {
		raws = append(raws, rawEntry{t: p.CreatedAt, Type: "Purchase", Ref: p.InvoiceNo, Debit: p.TotalAmount})
		if p.PaidAmount > 0 {
			raws = append(raws, rawEntry{t: p.CreatedAt, Type: "Payment (at purchase)", Ref: p.InvoiceNo, Credit: p.PaidAmount})
		}
	}
	for _, p := range payments {
		raws = append(raws, rawEntry{t: p.CreatedAt, Type: "Payment", Ref: p.Method, Credit: p.Amount})
	}
	for _, r := range returns {
		raws = append(raws, rawEntry{t: r.CreatedAt, Type: "Purchase Return", Ref: r.InvoiceNo, Credit: r.TotalAmount})
	}

	return build(supplier.ID, supplier.Name, raws), nil
}
