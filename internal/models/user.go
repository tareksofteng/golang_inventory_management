package models

// User is an authenticated operator of the system. Role is a plain string
// column validated against the rbac package's known roles at the service layer.
type User struct {
	BaseModel
	Name     string `gorm:"type:varchar(100);not null" json:"name"`
	Email    string `gorm:"type:varchar(100);not null;uniqueIndex" json:"email"`
	Password string `gorm:"type:varchar(255);not null" json:"-"` // bcrypt hash — json:"-" so it is NEVER serialized
	Role     string `gorm:"type:varchar(20);not null" json:"role"`
	IsActive bool   `gorm:"not null" json:"is_active"`
}
