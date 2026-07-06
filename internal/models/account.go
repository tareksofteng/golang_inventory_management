package models

// Account is one line of the Chart of Accounts. Type drives the "normal side":
// asset & expense are debit-normal; liability, equity & income are credit-normal.
type Account struct {
	BaseModel
	Code     string `gorm:"type:varchar(20);not null;uniqueIndex" json:"code"`
	Name     string `gorm:"type:varchar(100);not null" json:"name"`
	Type     string `gorm:"type:varchar(20);not null;index" json:"type"` // asset|liability|equity|income|expense
	IsActive bool   `gorm:"not null" json:"is_active"`
}
