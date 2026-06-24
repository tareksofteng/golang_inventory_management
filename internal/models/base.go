package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel holds the columns every table shares. We embed it into each model
// struct so we don't repeat these four fields everywhere.
//
// Laravel comparison:
//   - ID                       -> $table->id()
//   - CreatedAt / UpdatedAt    -> $table->timestamps()
//   - DeletedAt (gorm.DeletedAt) -> $table->softDeletes() + the SoftDeletes trait
//
// GORM recognises these field names by convention and fills them automatically:
//   - CreatedAt is set on insert
//   - UpdatedAt is set on every save
//   - DeletedAt enables SOFT DELETE: Delete() sets the timestamp instead of
//     removing the row, and every normal query auto-filters out soft-deleted
//     rows (WHERE deleted_at IS NULL).
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
