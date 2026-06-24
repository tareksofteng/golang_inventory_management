package models

// StockIn records a single stock-receiving event: "we received N units of
// product X from supplier Y at unit cost Z". It is an audit trail — one row
// per receipt — and is never edited after creation.
//
// Business rule (handled later in the service layer, inside a DB transaction):
// creating a StockIn must also INCREMENT the related Product.Quantity. Storing
// both the event log AND the running total is intentional: the log gives
// history, the Product.Quantity gives a fast current value without summing
// every StockIn row.
type StockIn struct {
	BaseModel

	ProductID  uint `gorm:"not null;index" json:"product_id"`
	SupplierID uint `gorm:"not null;index" json:"supplier_id"`

	Quantity int     `gorm:"not null" json:"quantity"`                               // units received (must be > 0)
	UnitCost float64 `gorm:"type:decimal(12,2);not null;default:0" json:"unit_cost"` // buying price per unit at this receipt
	Note     string  `gorm:"type:varchar(255)" json:"note"`                          // optional remark / invoice ref

	// Associations — populated only when preloaded.
	Product  *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Supplier *Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}
