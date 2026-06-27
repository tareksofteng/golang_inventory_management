package models

// Purchase is a stock-receiving invoice from a supplier. Creating one is a
// transactional operation: it inserts the invoice + items, increases each
// product's stock, and adds any unpaid amount to the supplier's due.
type Purchase struct {
	BaseModel
	InvoiceNo   string  `gorm:"type:varchar(30);not null;uniqueIndex" json:"invoice_no"`
	SupplierID  uint    `gorm:"not null;index" json:"supplier_id"`
	UserID      uint    `gorm:"not null;index" json:"user_id"` // who recorded it
	TotalAmount float64 `gorm:"type:decimal(14,2);not null" json:"total_amount"`
	PaidAmount  float64 `gorm:"type:decimal(14,2);not null" json:"paid_amount"`
	Due         float64 `gorm:"type:decimal(14,2);not null" json:"due"`
	Note        string  `gorm:"type:varchar(255)" json:"note"`

	Supplier *Supplier      `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Items    []PurchaseItem `gorm:"foreignKey:PurchaseID" json:"items,omitempty"`
}

// PurchaseItem is one line of a purchase invoice.
type PurchaseItem struct {
	BaseModel
	PurchaseID uint    `gorm:"not null;index" json:"purchase_id"`
	ProductID  uint    `gorm:"not null;index" json:"product_id"`
	Quantity   int     `gorm:"not null" json:"quantity"`
	UnitCost   float64 `gorm:"type:decimal(12,2);not null" json:"unit_cost"`
	Subtotal   float64 `gorm:"type:decimal(14,2);not null" json:"subtotal"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
