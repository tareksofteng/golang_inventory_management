// Package seeder inserts essential baseline data on startup.
package seeder

import (
	"log"

	"inventory-api/config"
	"inventory-api/internal/models"
	"inventory-api/internal/rbac"
	"inventory-api/pkg/auth"

	"gorm.io/gorm"
)

// SeedSuperAdmin creates the first super-admin IF no users exist yet. It is
// idempotent — safe to call on every startup. Without this there would be no
// account to log in with on a fresh database.
func SeedSuperAdmin(db *gorm.DB, cfg config.SeedConfig) error {
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // users already exist; nothing to seed
	}

	hash, err := auth.HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}

	admin := &models.User{
		Name:     "Super Admin",
		Email:    cfg.AdminEmail,
		Password: hash,
		Role:     string(rbac.RoleSuperAdmin),
		IsActive: true,
	}
	if err := db.Create(admin).Error; err != nil {
		return err
	}

	log.Printf("seeder: super-admin created (email=%s)", cfg.AdminEmail)
	return nil
}
