package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

// AuditLogRepository persists and reads the write-operation audit trail.
type AuditLogRepository interface {
	Create(entry *models.AuditLog) error
	List(offset, limit int) ([]models.AuditLog, int64, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(entry *models.AuditLog) error {
	return r.db.Create(entry).Error
}

// List returns the audit trail newest-first, with the acting user preloaded.
func (r *auditLogRepository) List(offset, limit int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	if err := r.db.Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.Preload("User").Order("id DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}
