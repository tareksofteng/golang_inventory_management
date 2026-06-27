package models

// Payment records money received from a customer or paid to a supplier. It is a
// polymorphic record: PartyType says which side, PartyID points at the row. We
// also snapshot PartyName so payment history reads well without extra joins.
type Payment struct {
	BaseModel
	PartyType string  `gorm:"type:varchar(10);not null;index" json:"party_type"` // "customer" | "supplier"
	PartyID   uint    `gorm:"not null;index" json:"party_id"`
	PartyName string  `gorm:"type:varchar(100)" json:"party_name"`
	UserID    uint    `gorm:"not null;index" json:"user_id"`
	Amount    float64 `gorm:"type:decimal(14,2);not null" json:"amount"`
	Method    string  `gorm:"type:varchar(20);not null" json:"method"` // cash | bank | mobile
	Note      string  `gorm:"type:varchar(255)" json:"note"`
}
