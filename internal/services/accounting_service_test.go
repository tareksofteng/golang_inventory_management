package services

import (
	"testing"
	"time"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"
)

// stubJournalRepo feeds canned per-account debit/credit totals into the trial
// balance; the rest of JournalRepository is unreachable here.
type stubJournalRepo struct {
	repositories.JournalRepository
	balances []repositories.AccountBalance
}

func (s stubJournalRepo) TrialBalance() ([]repositories.AccountBalance, error) {
	return s.balances, nil
}

func (s stubJournalRepo) TrialBalanceBetween(_, _ time.Time) ([]repositories.AccountBalance, error) {
	return s.balances, nil
}

// stubAccountRepo supplies the code/name/type lookup the trial balance joins on.
type stubAccountRepo struct {
	repositories.AccountRepository
	accounts []models.Account
}

func (s stubAccountRepo) FindAll(_, _ string, _, _ int) ([]models.Account, int64, error) {
	return s.accounts, int64(len(s.accounts)), nil
}

func acct(id uint, code, name, typ string) models.Account {
	a := models.Account{Code: code, Name: name, Type: typ}
	a.ID = id
	return a
}

// TrialBalance is the heart of the double-entry ledger: each account's debits and
// credits are netted into a SINGLE column, accounts that net to zero are dropped,
// and in a correct ledger the debit and credit totals match. These rules are
// where accounting bugs hide, so they are pinned explicitly here.
func TestTrialBalanceNetting(t *testing.T) {
	journal := stubJournalRepo{balances: []repositories.AccountBalance{
		{AccountID: 1, Debit: 15000, Credit: 5000}, // net +10000 -> debit column
		{AccountID: 2, Debit: 2000, Credit: 0},     // net +2000  -> debit column
		{AccountID: 3, Debit: 0, Credit: 12000},    // net -12000 -> credit column
		{AccountID: 4, Debit: 500, Credit: 500},    // nets to zero -> omitted
		{AccountID: 5, Debit: 10.004, Credit: 10},  // 0.004 < half a cent -> omitted
	}}
	accounts := stubAccountRepo{accounts: []models.Account{
		acct(1, "1000", "Cash", "asset"),
		acct(2, "1100", "Accounts Receivable", "asset"),
		acct(3, "4000", "Sales Income", "income"),
		acct(4, "5000", "Rounding Wash", "expense"),
		acct(5, "5100", "Dust", "expense"),
	}}

	svc := NewAccountingService(journal, accounts)

	tb, err := svc.TrialBalance()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Zero-net and sub-cent accounts are dropped: 5 inputs -> 3 rows.
	if len(tb.Rows) != 3 {
		t.Fatalf("got %d rows, want 3 (zero-net rows must be omitted)", len(tb.Rows))
	}

	rows := make(map[uint]TrialBalanceRow, len(tb.Rows))
	for _, r := range tb.Rows {
		rows[r.AccountID] = r
	}

	// A debit-heavy account lands entirely in the debit column, metadata joined.
	cash := rows[1]
	assertMoney(t, "cash debit", cash.Debit, 10000)
	assertMoney(t, "cash credit", cash.Credit, 0)
	if cash.Code != "1000" || cash.Type != "asset" {
		t.Errorf("cash metadata not joined: %+v", cash)
	}

	// A credit-heavy account lands in the credit column.
	sales := rows[3]
	assertMoney(t, "sales credit", sales.Credit, 12000)
	assertMoney(t, "sales debit", sales.Debit, 0)

	// The omitted accounts must not appear at all.
	if _, ok := rows[4]; ok {
		t.Error("zero-net account 4 should have been omitted")
	}
	if _, ok := rows[5]; ok {
		t.Error("sub-cent account 5 should have been omitted")
	}

	// Debits equal credits — the ledger balances (12000 == 12000).
	assertMoney(t, "total debit", tb.TotalDebit, 12000)
	assertMoney(t, "total credit", tb.TotalCredit, 12000)
	if tb.TotalDebit != tb.TotalCredit {
		t.Errorf("trial balance does not balance: debit %.2f != credit %.2f", tb.TotalDebit, tb.TotalCredit)
	}
}

// Profit & Loss layers three totals: gross profit = income - COGS, then net
// profit = gross profit - operating expenses. COGS (account 5000) must be split
// out from ordinary expenses, and balance-sheet accounts (assets, equity) must
// never leak into the income statement.
func TestProfitLossComputation(t *testing.T) {
	journal := stubJournalRepo{balances: []repositories.AccountBalance{
		{AccountID: 1, Debit: 0, Credit: 20000}, // Sales Income
		{AccountID: 2, Debit: 8000, Credit: 0},  // COGS (code 5000)
		{AccountID: 3, Debit: 3000, Credit: 0},  // Salary Expense
		{AccountID: 4, Debit: 50000, Credit: 0}, // Cash (asset) — not a P&L line
		{AccountID: 5, Debit: 100, Credit: 100}, // income netting to zero — dropped
	}}
	accounts := stubAccountRepo{accounts: []models.Account{
		acct(1, "4000", "Sales Income", "income"),
		acct(2, "5000", "Cost of Goods Sold", "expense"),
		acct(3, "5100", "Salary Expense", "expense"),
		acct(4, "1000", "Cash", "asset"),
		acct(5, "4100", "Misc Income", "income"),
	}}

	svc := NewAccountingService(journal, accounts)

	pl, err := svc.ProfitLoss(time.Now(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pl.Income) != 1 || len(pl.COGS) != 1 || len(pl.Expenses) != 1 {
		t.Fatalf("section sizes income=%d cogs=%d expenses=%d, want 1/1/1 (asset + zero-net dropped)",
			len(pl.Income), len(pl.COGS), len(pl.Expenses))
	}

	assertMoney(t, "total income", pl.TotalIncome, 20000)
	assertMoney(t, "total COGS", pl.TotalCOGS, 8000)
	assertMoney(t, "total expense", pl.TotalExpense, 3000)
	assertMoney(t, "gross profit", pl.GrossProfit, 12000) // 20000 - 8000
	assertMoney(t, "net profit", pl.NetProfit, 9000)      // 12000 - 3000
}
