// Package auth provides authentication primitives: password hashing, JWT
// access tokens, and opaque refresh tokens. It is policy-free (knows nothing
// about roles/permissions — that is the rbac package's job).
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidToken is returned for any malformed, tampered, or expired token.
var ErrInvalidToken = errors.New("invalid or expired token")

// ---- Password hashing (bcrypt) --------------------------------------------

// HashPassword returns a bcrypt hash of the plaintext password.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword reports whether plain matches the stored bcrypt hash.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// ---- JWT access tokens -----------------------------------------------------

// Claims is the payload carried inside an access token. uid + role let the
// auth/RBAC middleware authorize a request WITHOUT a DB lookup.
type Claims struct {
	UserID      uint     `json:"uid"`
	Role        string   `json:"role"`
	Permissions []string `json:"perms"`
	jwt.RegisteredClaims
}

// TokenManager signs and verifies access tokens with one secret + TTLs.
type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenManager builds a TokenManager from config values.
func NewTokenManager(secret string, accessTTLMinutes, refreshTTLHours int) *TokenManager {
	return &TokenManager{
		secret:     []byte(secret),
		accessTTL:  time.Duration(accessTTLMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshTTLHours) * time.Hour,
	}
}

// RefreshTTL exposes the refresh-token lifetime so the service can set the DB
// expiry consistently.
func (tm *TokenManager) RefreshTTL() time.Duration { return tm.refreshTTL }

// GenerateAccessToken signs a short-lived JWT for the given user, role and
// effective permissions.
func (tm *TokenManager) GenerateAccessToken(userID uint, role string, permissions []string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:      userID,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secret)
}

// ParseAccessToken verifies the signature + expiry and returns the claims.
func (tm *TokenManager) ParseAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		// Reject any token not signed with HMAC — defends against the classic
		// "alg: none" / algorithm-confusion attack.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return tm.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ---- Refresh tokens (opaque, stored hashed) --------------------------------

// GenerateRefreshToken returns a cryptographically-random 256-bit token (hex).
// This raw value is given to the client ONCE; only its hash is stored.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashRefreshToken returns the sha256 hex of a refresh token, for safe storage
// and constant-shape lookups. (sha256 is fine here — the token already has full
// entropy, unlike a human password which needs bcrypt's slowness.)
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
