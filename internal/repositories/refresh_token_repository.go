package repositories

import (
	"inventory-api/internal/models"

	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByHash(hash string) (*models.RefreshToken, error)
	DeleteByHash(hash string) error
	DeleteByUserID(userID uint) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindByHash(hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.Where("token_hash = ?", hash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteByHash permanently removes one refresh token (logout of one session).
// We HARD-delete here (Unscoped): a revoked token must truly be gone, not just
// soft-deleted where it could still be matched.
func (r *refreshTokenRepository) DeleteByHash(hash string) error {
	return r.db.Unscoped().Where("token_hash = ?", hash).Delete(&models.RefreshToken{}).Error
}

// DeleteByUserID removes all of a user's refresh tokens (e.g. on disable or
// "logout everywhere").
func (r *refreshTokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Unscoped().Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}
