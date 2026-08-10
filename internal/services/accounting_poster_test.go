package services

import (
	"math"
	"testing"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"
)

// Account ids the stub hands back, keyed by the code the poster asks for.
var stubAccountIDs = map[string]uint{
	accCash:      1,
	accAR:        2,
	accInventory: 3,
	accAP:        4,
	accSales:     5,
	accCOGS:      6,
}

// stubAccounts embeds the interface so it satisfies AccountRepository without
// implementing every method. Only FindByCode is reachable from the poster; any
// other call would panic, which is the behaviour we want in a test.
type stubAccounts struct {
	repositories.AccountRepository
}

func (stubAccounts) FindByCode(code string) (*models.Account, error) {
	acc := &models.Account{Code: code}
	acc.ID = stubAccountIDs[code]
	return acc, nil
}

// recordingJournal captures what the poster tried to write instead of hitting
// the database.
type recordingJournal struct {
	JournalService
	entries []CreateJournalInput
}

func (r *recordingJournal) Create(input CreateJournalInput) (*models.JournalEntry, error) {
	r.entries = append(r.entries, input)
	return &models.JournalEntry{}, nil
}

func newTestPoster() (AccountingPoster, *recordingJournal) {
	journal := &recordingJournal{}
	return NewAccountingPoster(journal, stubAccounts{}), journal
}

// Every auto-posted entry must balance. If this ever fails, the trial balance
// and both financial statements are wrong for that transaction.
func TestAutoPostedEntriesBalance(t *testing.T) {
	sale := &models.Sale{
		InvoiceNo:   "INV-000001",
		UserID:      1,
		TotalAmount: 1190,
		PaidAmount:  500,
		Due:         690,
		Items: []models.SaleItem{
			{Quantity: 2, UnitPrice: 500, UnitCost: 300},
			{Quantity: 1, UnitPrice: 190, UnitCost: 120},
		},
	}
	purchase := &models.Purchase{
		InvoiceNo:   "PUR-000001",
		UserID:      1,
		TotalAmount: 800,
		PaidAmount:  300,
		Due:         500,
	}
	receipt := &models.Payment{PartyType: "customer", PartyName: "Acme", UserID: 1, Amount: 690}
	outgoing := &models.Payment{PartyType: "supplier", PartyName: "Supplier Co", UserID: 1, Amount: 500}
	saleReturn := &models.SaleReturn{
		InvoiceNo:   "SR-000001",
		UserID:      1,
		TotalAmount: 190,
		Items:       []models.SaleReturnItem{{Quantity: 1, UnitCost: 120}},
	}
	purchaseReturn := &models.PurchaseReturn{InvoiceNo: "PR-000001", UserID: 1, TotalAmount: 200}

	cases := []struct {
		name string
		post func(AccountingPoster)
	}{
		{"sale", func(p AccountingPoster) { p.PostSale(sale) }},
		{"purchase", func(p AccountingPoster) { p.PostPurchase(purchase) }},
		{"customer payment", func(p AccountingPoster) { p.PostCustomerPayment(receipt) }},
		{"supplier payment", func(p AccountingPoster) { p.PostSupplierPayment(outgoing) }},
		{"sale return", func(p AccountingPoster) { p.PostSaleReturn(saleReturn) }},
		{"purchase return", func(p AccountingPoster) { p.PostPurchaseReturn(purchaseReturn) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			poster, journal := newTestPoster()
			tc.post(poster)

			if len(journal.entries) != 1 {
				t.Fatalf("expected exactly 1 journal entry, got %d", len(journal.entries))
			}
			entry := journal.entries[0]
			if len(entry.Lines) < 2 {
				t.Fatalf("a double-entry posting needs at least 2 lines, got %d", len(entry.Lines))
			}

			var debit, credit float64
			for _, line := range entry.Lines {
				if line.Debit > 0 && line.Credit > 0 {
					t.Errorf("line on account %d has both a debit and a credit", line.AccountID)
				}
				if line.AccountID == 0 {
					t.Errorf("line posted to an unresolved account (id 0)")
				}
				debit += line.Debit
				credit += line.Credit
			}
			if math.Abs(debit-credit) > 0.005 {
				t.Errorf("entry does not balance: debit %v, credit %v", debit, credit)
			}
		})
	}
}

// A credit sale splits the debit between cash received and the amount still
// owed, and moves the goods' cost out of inventory into COGS.
func TestPostSaleSplitsCashAndReceivableAndMovesCost(t *testing.T) {
	poster, journal := newTestPoster()
	poster.PostSale(&models.Sale{
		InvoiceNo:   "INV-000002",
		UserID:      1,
		TotalAmount: 1000,
		PaidAmount:  400,
		Due:         600,
		Items:       []models.SaleItem{{Quantity: 3, UnitPrice: 200, UnitCost: 150}},
	})

	entry := journal.entries[0]
	const cost = 3 * 150.0

	assertMoney(t, "cash debit", debitOn(entry, stubAccountIDs[accCash]), 400)
	assertMoney(t, "receivable debit", debitOn(entry, stubAccountIDs[accAR]), 600)
	assertMoney(t, "sales credit", creditOn(entry, stubAccountIDs[accSales]), 1000)
	assertMoney(t, "cogs debit", debitOn(entry, stubAccountIDs[accCOGS]), cost)
	assertMoney(t, "inventory credit", creditOn(entry, stubAccountIDs[accInventory]), cost)
}

// A fully paid sale must not open a receivable.
func TestPostSaleWithoutDueSkipsReceivable(t *testing.T) {
	poster, journal := newTestPoster()
	poster.PostSale(&models.Sale{
		InvoiceNo:   "INV-000003",
		UserID:      1,
		TotalAmount: 250,
		PaidAmount:  250,
		Due:         0,
		Items:       []models.SaleItem{{Quantity: 1, UnitPrice: 250, UnitCost: 100}},
	})

	if got := debitOn(journal.entries[0], stubAccountIDs[accAR]); got != 0 {
		t.Errorf("a fully paid sale posted %v to receivables", got)
	}
}

// Posting is best-effort and must stay silent on empty transactions rather
// than writing a zero-value entry into the ledger.
func TestZeroValueTransactionsPostNothing(t *testing.T) {
	cases := []struct {
		name string
		post func(AccountingPoster)
	}{
		{"sale", func(p AccountingPoster) { p.PostSale(&models.Sale{TotalAmount: 0}) }},
		{"purchase", func(p AccountingPoster) { p.PostPurchase(&models.Purchase{TotalAmount: 0}) }},
		{"customer payment", func(p AccountingPoster) { p.PostCustomerPayment(&models.Payment{Amount: 0}) }},
		{"supplier payment", func(p AccountingPoster) { p.PostSupplierPayment(&models.Payment{Amount: -5}) }},
		{"sale return", func(p AccountingPoster) { p.PostSaleReturn(&models.SaleReturn{TotalAmount: 0}) }},
		{"purchase return", func(p AccountingPoster) { p.PostPurchaseReturn(&models.PurchaseReturn{TotalAmount: 0}) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			poster, journal := newTestPoster()
			tc.post(poster)
			if len(journal.entries) != 0 {
				t.Errorf("expected no journal entry, got %d", len(journal.entries))
			}
		})
	}
}

func debitOn(entry CreateJournalInput, accountID uint) float64 {
	var sum float64
	for _, line := range entry.Lines {
		if line.AccountID == accountID {
			sum += line.Debit
		}
	}
	return sum
}

func creditOn(entry CreateJournalInput, accountID uint) float64 {
	var sum float64
	for _, line := range entry.Lines {
		if line.AccountID == accountID {
			sum += line.Credit
		}
	}
	return sum
}
