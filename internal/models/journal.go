package models

import "time"

// JournalEntry is one balanced double-entry transaction (total debit == total
// credit across its lines).
type JournalEntry struct {
	BaseModel
	EntryNo   string    `gorm:"type:varchar(30);not null;uniqueIndex" json:"entry_no"`
	Date      time.Time `gorm:"not null;index" json:"date"`
	Reference string    `gorm:"type:varchar(100)" json:"reference"`
	Note      string    `gorm:"type:varchar(255)" json:"note"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`

	Lines []JournalLine `gorm:"foreignKey:JournalEntryID" json:"lines,omitempty"`
}

// JournalLine posts a debit or a credit to one account. Exactly one of
// Debit/Credit is non-zero.
type JournalLine struct {
	BaseModel
	JournalEntryID uint    `gorm:"not null;index" json:"journal_entry_id"`
	AccountID      uint    `gorm:"not null;index" json:"account_id"`
	Debit          float64 `gorm:"type:decimal(14,2);not null" json:"debit"`
	Credit         float64 `gorm:"type:decimal(14,2);not null" json:"credit"`

	Account *Account `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}
