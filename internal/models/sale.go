package models

// Sale is a selling invoice to a customer. Creating one is transactional: it
// inserts the invoice + items, decreases each product's stock (failing if there
// is not enough), and adds any unpaid amount to the customer's due.
type Sale struct {
	BaseModel
	InvoiceNo   string  `gorm:"type:varchar(30);not null;uniqueIndex" json:"invoice_no"`
	CustomerID  uint    `gorm:"not null;index" json:"customer_id"`
	UserID      uint    `gorm:"not null;index" json:"user_id"`
	TotalAmount float64 `gorm:"type:decimal(14,2);not null" json:"total_amount"`
	PaidAmount  float64 `gorm:"type:decimal(14,2);not null" json:"paid_amount"`
	Due         float64 `gorm:"type:decimal(14,2);not null" json:"due"`
	Note        string  `gorm:"type:varchar(255)" json:"note"`

	Customer *Customer  `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Items    []SaleItem `gorm:"foreignKey:SaleID" json:"items,omitempty"`
}

// SaleItem is one line of a sale invoice. UnitPrice is the SELLING price.
type SaleItem struct {
	BaseModel
	SaleID    uint    `gorm:"not null;index" json:"sale_id"`
	ProductID uint    `gorm:"not null;index" json:"product_id"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	UnitPrice float64 `gorm:"type:decimal(12,2);not null" json:"unit_price"`
	Subtotal  float64 `gorm:"type:decimal(14,2);not null" json:"subtotal"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
