package models

// Category groups products (e.g. "Electronics", "Groceries").
//
// We embed BaseModel to inherit ID, CreatedAt, UpdatedAt and the soft-delete
// DeletedAt column. Only the Category-specific fields live here.
//
// The `gorm` tags describe the DB column; the `json` tags describe the API
// shape. Keeping both on one struct is fine for a project this size — we
// validate incoming data with separate request structs (DTOs) at the
// controller layer, NOT with binding tags here, so the model stays a pure
// data/persistence type.
type Category struct {
	BaseModel
	Name        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`
	// No `default:true` here on purpose: GORM would then drop a false zero-value
	// on insert and force true. Defaulting lives in the controller instead, so
	// an explicit is_active:false is honoured.
	IsActive bool `gorm:"not null" json:"is_active"`
}
