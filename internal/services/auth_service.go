package services

import (
	"errors"
	"time"

	"inventory-api/internal/models"
	"inventory-api/internal/repositories"
	"inventory-api/pkg/auth"

	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserDisabled       = errors.New("user account is disabled")
)

// AuthService handles login, token refresh, and logout.
type AuthService interface {
	Login(email, password string) (user *models.User, accessToken, refreshToken string, err error)
	Refresh(refreshToken string) (accessToken, newRefreshToken string, err error)
	Logout(refreshToken string) error
}

type authService struct {
	userRepo    repositories.UserRepository
	refreshRepo repositories.RefreshTokenRepository
	tokens      *auth.TokenManager
}

func NewAuthService(
	userRepo repositories.UserRepository,
	refreshRepo repositories.RefreshTokenRepository,
	tokens *auth.TokenManager,
) AuthService {
	return &authService{userRepo: userRepo, refreshRepo: refreshRepo, tokens: tokens}
}

// Login verifies credentials and issues an access + refresh token pair.
func (s *authService) Login(email, password string) (*models.User, string, string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Same error for "no such email" and "wrong password" so an
			// attacker cannot tell which emails exist (user enumeration).
			return nil, "", "", ErrInvalidCredentials
		}
		return nil, "", "", err
	}
	if !user.IsActive {
		return nil, "", "", ErrUserDisabled
	}
	if !auth.CheckPassword(user.Password, password) {
		return nil, "", "", ErrInvalidCredentials
	}

	access, refresh, err := s.issueTokens(user)
	if err != nil {
		return nil, "", "", err
	}
	return user, access, refresh, nil
}

// Refresh validates a refresh token, rotates it, and returns a fresh pair.
func (s *authService) Refresh(refreshToken string) (string, string, error) {
	hash := auth.HashRefreshToken(refreshToken)
	stored, err := s.refreshRepo.FindByHash(hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", auth.ErrInvalidToken
		}
		return "", "", err
	}

	// Expired? remove it and reject.
	if time.Now().After(stored.ExpiresAt) {
		_ = s.refreshRepo.DeleteByHash(hash)
		return "", "", auth.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(stored.UserID)
	if err != nil {
		return "", "", err
	}
	if !user.IsActive {
		return "", "", ErrUserDisabled
	}

	// Rotation: invalidate the used refresh token, issue a brand new pair.
	// If a stolen token is replayed after rotation, lookup fails -> rejected.
	if err := s.refreshRepo.DeleteByHash(hash); err != nil {
		return "", "", err
	}
	return s.issueTokens(user)
}

// Logout revokes a single refresh token. Unknown tokens are treated as success
// (idempotent logout).
func (s *authService) Logout(refreshToken string) error {
	return s.refreshRepo.DeleteByHash(auth.HashRefreshToken(refreshToken))
}

// issueTokens generates an access JWT plus a persisted (hashed) refresh token.
func (s *authService) issueTokens(user *models.User) (string, string, error) {
	access, err := s.tokens.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", err
	}

	refresh, err := auth.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	record := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: auth.HashRefreshToken(refresh),
		ExpiresAt: time.Now().Add(s.tokens.RefreshTTL()),
	}
	if err := s.refreshRepo.Create(record); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}
