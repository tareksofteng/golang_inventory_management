// Package middleware holds Gin middleware: authentication (verify the JWT) and
// authorization (enforce RBAC permissions).
package middleware

import (
	"net/http"
	"strings"

	"inventory-api/internal/rbac"
	"inventory-api/pkg/auth"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// Context keys under which the authenticated user's id and role are stored.
const (
	ctxUserID   = "auth_user_id"
	ctxUserRole = "auth_user_role"
)

// Auth verifies the "Authorization: Bearer <token>" header. On success it puts
// the user id + role into the Gin context for downstream handlers; on failure
// it aborts with 401. This is the gatekeeper for every protected route.
func Auth(tm *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, "missing or malformed Authorization header", nil)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := tm.ParseAccessToken(tokenStr)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token", nil)
			c.Abort()
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUserRole, claims.Role)
		c.Next()
	}
}

// RequirePermission ensures the authenticated user's role grants perm, else 403.
// It must run AFTER Auth (which sets the role in context).
func RequirePermission(perm rbac.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rbac.HasPermission(rbac.Role(UserRole(c)), perm) {
			response.Error(c, http.StatusForbidden, "you do not have permission to perform this action", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserID returns the authenticated user's id from context (0 if unauthenticated).
func UserID(c *gin.Context) uint {
	v, _ := c.Get(ctxUserID)
	id, _ := v.(uint)
	return id
}

// UserRole returns the authenticated user's role from context ("" if none).
func UserRole(c *gin.Context) string {
	v, _ := c.Get(ctxUserRole)
	s, _ := v.(string)
	return s
}
