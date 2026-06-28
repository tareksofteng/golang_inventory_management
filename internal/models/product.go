package models

// Product is a sellable/stockable item. It belongs to one Category and is
// supplied by one Supplier.
//
// Relationships in GORM are expressed with two pieces:
//  1. A foreign-key column field  -> CategoryID / SupplierID (the actual DB column)
//  2. An association struct field  -> Category / Supplier (filled only when you
//     explicitly Preload it; like Laravel's $product->load('category'))
//
// Laravel comparison: where Eloquent has `belongsTo(Category::class)`, GORM
// infers the relation from the `CategoryID` naming convention + the foreignKey
// tag. No separate relationship method is needed.
type Product struct {
	BaseModel

	Name  string `gorm:"type:varchar(150);not null" json:"name"`
	SKU   string `gorm:"type:varchar(50);not null;uniqueIndex" json:"sku"` // Stock Keeping Unit — unique product code
	Image string `gorm:"type:varchar(255)" json:"image"`                   // relative URL e.g. /uploads/abc.jpg

	// Foreign keys. `index` speeds up the JOIN/filter on these columns.
	CategoryID uint `gorm:"not null;index" json:"category_id"`
	SupplierID uint `gorm:"not null;index" json:"supplier_id"`

	// Money is stored as DECIMAL(12,2) for exact values (never FLOAT in DB).
	Price     float64 `gorm:"type:decimal(12,2);not null;default:0" json:"price"`      // selling price
	CostPrice float64 `gorm:"type:decimal(12,2);not null;default:0" json:"cost_price"` // buying price

	Quantity int    `gorm:"not null;default:0" json:"quantity"` // current stock on hand
	Unit     string `gorm:"type:varchar(20)" json:"unit"`       // pcs, kg, litre...
	IsActive bool   `gorm:"not null" json:"is_active"`          // defaulted in controller, not via GORM (false zero-value gotcha)

	// Associations. Pointers + omitempty so they vanish from the JSON when not
	// preloaded, instead of showing an empty {} object.
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Supplier *Supplier `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}
