package models

// StockAdjustment records a manual correction to a product's stock that did not
// come from a purchase or a sale — damaged goods, theft/loss, a physical
// recount, or stock found. It is an append-only audit trail: one row per
// correction, never edited.
//
// Business rule (handled in the repository inside a DB transaction): creating a
// StockAdjustment also moves Product.Quantity by Change. NewQuantity stores the
// resulting stock so the history reads correctly even after later movements.
type StockAdjustment struct {
	BaseModel

	ProductID uint `gorm:"not null;index" json:"product_id"`
	UserID    uint `gorm:"not null;index" json:"user_id"`

	Change      int    `gorm:"not null" json:"change"`                  // +N added, -N removed
	NewQuantity int    `gorm:"not null" json:"new_quantity"`            // product stock AFTER this adjustment
	Reason      string `gorm:"type:varchar(30);not null" json:"reason"` // damaged, lost, recount, found, other
	Note        string `gorm:"type:varchar(255)" json:"note"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
