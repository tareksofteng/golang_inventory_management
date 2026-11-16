package models

// AuditLog is an append-only record of every data-changing request (create,
// update, delete). It exists for traceability/accountability: who did what and
// when. Read requests are never logged. Rows are written by the audit
// middleware and are never edited or deleted by the application.
type AuditLog struct {
	BaseModel

	UserID uint   `gorm:"index" json:"user_id"`           // the authenticated user who made the change
	Method string `gorm:"type:varchar(10)" json:"method"` // POST, PUT, PATCH, DELETE
	Path   string `gorm:"type:varchar(255)" json:"path"`  // the matched route pattern, e.g. /api/v1/products/:id
	Status int    `gorm:"not null" json:"status"`         // HTTP status the request returned
	IP     string `gorm:"type:varchar(45)" json:"ip"`     // client IP (IPv4/IPv6)

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
