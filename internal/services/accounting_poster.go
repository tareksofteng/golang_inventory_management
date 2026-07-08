package services

import (
	"log"
	"time"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"
)

// Standard account codes used for auto-posting (from the seeded chart).
const (
	accCash      = "1000"
	accAR        = "1100" // Accounts Receivable
	accInventory = "1200"
	accAP        = "2000" // Accounts Payable
	accSales     = "4000" // Sales Income
	accCOGS      = "5000" // Cost of Goods Sold
)

// AccountingPoster turns a business transaction into a balanced journal entry.
// Posting is best-effort: a failure is logged but does not roll back the
// business transaction (which is already committed).
type AccountingPoster interface {
	PostSale(sale *models.Sale)
	PostPurchase(p *models.Purchase)
	PostCustomerPayment(p *models.Payment)
	PostSupplierPayment(p *models.Payment)
	PostSaleReturn(r *models.SaleReturn)
	PostPurchaseReturn(r *models.PurchaseReturn)
}

type accountingPoster struct {
	journal  JournalService
	accounts repositories.AccountRepository
}

func NewAccountingPoster(journal JournalService, accounts repositories.AccountRepository) AccountingPoster {
	return &accountingPoster{journal: journal, accounts: accounts}
}

// id resolves an account code to its id (0 if missing).
func (p *accountingPoster) id(code string) uint {
	a, err := p.accounts.FindByCode(code)
	if err != nil {
		return 0
	}
	return a.ID
}

func (p *accountingPoster) post(date time.Time, ref, note string, userID uint, lines []JournalLineInput) {
	if len(lines) < 2 {
		return
	}
	if _, err := p.journal.Create(CreateJournalInput{Date: date, Reference: ref, Note: note, UserID: userID, Lines: lines}); err != nil {
		log.Printf("accounting: auto-post failed for %s: %v", ref, err)
	}
}

func dr(account uint, amount float64) JournalLineInput {
	return JournalLineInput{AccountID: account, Debit: amount}
}
func cr(account uint, amount float64) JournalLineInput {
	return JournalLineInput{AccountID: account, Credit: amount}
}

// Sale: Dr Cash (paid) + Dr A/R (due), Cr Sales Income (total).
func (p *accountingPoster) PostSale(s *models.Sale) {
	if s.TotalAmount <= 0 {
		return
	}
	var lines []JournalLineInput
	if s.PaidAmount > 0 {
		lines = append(lines, dr(p.id(accCash), s.PaidAmount))
	}
	if s.Due > 0 {
		lines = append(lines, dr(p.id(accAR), s.Due))
	}
	lines = append(lines, cr(p.id(accSales), s.TotalAmount))

	// COGS: move the goods' cost from Inventory to expense.
	var cost float64
	for _, it := range s.Items {
		cost += float64(it.Quantity) * it.UnitCost
	}
	if cost > 0 {
		lines = append(lines, dr(p.id(accCOGS), cost), cr(p.id(accInventory), cost))
	}
	p.post(s.CreatedAt, s.InvoiceNo, "Auto: Sale", s.UserID, lines)
}

// Purchase: Dr Inventory (total), Cr Cash (paid) + Cr A/P (due).
func (p *accountingPoster) PostPurchase(pu *models.Purchase) {
	if pu.TotalAmount <= 0 {
		return
	}
	lines := []JournalLineInput{dr(p.id(accInventory), pu.TotalAmount)}
	if pu.PaidAmount > 0 {
		lines = append(lines, cr(p.id(accCash), pu.PaidAmount))
	}
	if pu.Due > 0 {
		lines = append(lines, cr(p.id(accAP), pu.Due))
	}
	p.post(pu.CreatedAt, pu.InvoiceNo, "Auto: Purchase", pu.UserID, lines)
}

// Customer payment (receipt): Dr Cash, Cr A/R.
func (p *accountingPoster) PostCustomerPayment(pm *models.Payment) {
	if pm.Amount <= 0 {
		return
	}
	p.post(pm.CreatedAt, "Receipt "+pm.PartyName, "Auto: Customer payment", pm.UserID,
		[]JournalLineInput{dr(p.id(accCash), pm.Amount), cr(p.id(accAR), pm.Amount)})
}

// Supplier payment: Dr A/P, Cr Cash.
func (p *accountingPoster) PostSupplierPayment(pm *models.Payment) {
	if pm.Amount <= 0 {
		return
	}
	p.post(pm.CreatedAt, "Payment "+pm.PartyName, "Auto: Supplier payment", pm.UserID,
		[]JournalLineInput{dr(p.id(accAP), pm.Amount), cr(p.id(accCash), pm.Amount)})
}

// Sale return: reverses revenue — Dr Sales Income, Cr A/R.
func (p *accountingPoster) PostSaleReturn(r *models.SaleReturn) {
	if r.TotalAmount <= 0 {
		return
	}
	lines := []JournalLineInput{dr(p.id(accSales), r.TotalAmount), cr(p.id(accAR), r.TotalAmount)}
	// Reverse COGS: goods come back into Inventory.
	var cost float64
	for _, it := range r.Items {
		cost += float64(it.Quantity) * it.UnitCost
	}
	if cost > 0 {
		lines = append(lines, dr(p.id(accInventory), cost), cr(p.id(accCOGS), cost))
	}
	p.post(r.CreatedAt, r.InvoiceNo, "Auto: Sale return", r.UserID, lines)
}

// Purchase return: Dr A/P, Cr Inventory.
func (p *accountingPoster) PostPurchaseReturn(r *models.PurchaseReturn) {
	if r.TotalAmount <= 0 {
		return
	}
	p.post(r.CreatedAt, r.InvoiceNo, "Auto: Purchase return", r.UserID,
		[]JournalLineInput{dr(p.id(accAP), r.TotalAmount), cr(p.id(accInventory), r.TotalAmount)})
}
