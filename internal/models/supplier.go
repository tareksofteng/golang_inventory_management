package models

// Supplier is a company/person we buy stock from. Unlike Category, a supplier
// needs real-world contact details so the business can reach them.
type Supplier struct {
	BaseModel
	Name     string  `gorm:"type:varchar(100);not null" json:"name"`
	Email    string  `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	Phone    string  `gorm:"type:varchar(20)" json:"phone"`
	Address  string  `gorm:"type:varchar(255)" json:"address"`
	Due      float64 `gorm:"type:decimal(14,2);not null" json:"due"` // how much WE owe this supplier (grows on credit purchase)
	IsActive bool    `gorm:"not null" json:"is_active"`              // defaulted in controller, not via GORM (false zero-value gotcha)
}
