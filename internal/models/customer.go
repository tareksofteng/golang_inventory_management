package models

// Customer is a buyer. Unlike Supplier we do NOT make email unique: walk-in
// customers often share or lack an email, so a unique index would be wrong.
// Due is the customer's outstanding balance (how much they owe us); it starts
// at the opening balance and will later be moved by sales + payments.
type Customer struct {
	BaseModel
	Name     string  `gorm:"type:varchar(100);not null" json:"name"`
	Email    string  `gorm:"type:varchar(100)" json:"email"`
	Phone    string  `gorm:"type:varchar(20);index" json:"phone"`
	Address  string  `gorm:"type:varchar(255)" json:"address"`
	Due      float64 `gorm:"type:decimal(12,2);not null" json:"due"`
	IsActive bool    `gorm:"not null" json:"is_active"`
}
