package models

import "time"

// RefreshToken stores a HASH of an issued refresh token so the server can
// revoke it (logout) and verify it on refresh. We never store the raw token —
// only its sha256 hash, the same way we never store raw passwords.
type RefreshToken struct {
	BaseModel
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	TokenHash string    `gorm:"type:varchar(64);not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}
