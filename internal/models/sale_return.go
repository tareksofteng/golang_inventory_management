package models

// SaleReturn records goods a customer brought BACK. It is transactional: it
// increases stock (the goods are back on the shelf) and decreases the
// customer's due (they owe us less).
type SaleReturn struct {
	BaseModel
	InvoiceNo   string  `gorm:"type:varchar(30);not null;uniqueIndex" json:"invoice_no"`
	SaleID      uint    `gorm:"not null;index" json:"sale_id"` // the original sale being returned against
	CustomerID  uint    `gorm:"not null;index" json:"customer_id"`
	UserID      uint    `gorm:"not null;index" json:"user_id"`
	TotalAmount float64 `gorm:"type:decimal(14,2);not null" json:"total_amount"`
	Note        string  `gorm:"type:varchar(255)" json:"note"`

	Customer *Customer        `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Items    []SaleReturnItem `gorm:"foreignKey:SaleReturnID" json:"items,omitempty"`
}

type SaleReturnItem struct {
	BaseModel
	SaleReturnID uint    `gorm:"not null;index" json:"sale_return_id"`
	ProductID    uint    `gorm:"not null;index" json:"product_id"`
	Quantity     int     `gorm:"not null" json:"quantity"`
	UnitPrice    float64 `gorm:"type:decimal(12,2);not null" json:"unit_price"`
	Subtotal     float64 `gorm:"type:decimal(14,2);not null" json:"subtotal"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
