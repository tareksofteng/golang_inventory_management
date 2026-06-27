package models

// PurchaseReturn records goods sent BACK to a supplier. It is transactional:
// it decreases stock (we no longer hold the goods) and decreases the supplier's
// due (we owe them less). Cannot return more than is currently in stock.
type PurchaseReturn struct {
	BaseModel
	InvoiceNo   string  `gorm:"type:varchar(30);not null;uniqueIndex" json:"invoice_no"`
	SupplierID  uint    `gorm:"not null;index" json:"supplier_id"`
	UserID      uint    `gorm:"not null;index" json:"user_id"`
	TotalAmount float64 `gorm:"type:decimal(14,2);not null" json:"total_amount"`
	Note        string  `gorm:"type:varchar(255)" json:"note"`

	Supplier *Supplier            `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Items    []PurchaseReturnItem `gorm:"foreignKey:PurchaseReturnID" json:"items,omitempty"`
}

type PurchaseReturnItem struct {
	BaseModel
	PurchaseReturnID uint    `gorm:"not null;index" json:"purchase_return_id"`
	ProductID        uint    `gorm:"not null;index" json:"product_id"`
	Quantity         int     `gorm:"not null" json:"quantity"`
	UnitCost         float64 `gorm:"type:decimal(12,2);not null" json:"unit_cost"`
	Subtotal         float64 `gorm:"type:decimal(14,2);not null" json:"subtotal"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
