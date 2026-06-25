package controllers

import (
	"errors"
	"net/http"

	"inventory-api/internal/middleware"
	"inventory-api/internal/rbac"
	"inventory-api/internal/services"
	"inventory-api/pkg/auth"
	"inventory-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthController struct {
	authService services.AuthService
	userService services.UserService
}

func NewAuthController(authService services.AuthService, userService services.UserService) *AuthController {
	return &AuthController{authService: authService, userService: userService}
}

// Login -> POST /auth/login
func (ctrl *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	user, access, refresh, err := ctrl.authService.Login(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			response.Error(c, http.StatusUnauthorized, "Invalid email or password", nil)
		case errors.Is(err, services.ErrUserDisabled):
			response.Error(c, http.StatusForbidden, "Your account is disabled", nil)
		default:
			response.InternalError(c, "Login failed")
		}
		return
	}

	response.Success(c, "Login successful", gin.H{
		"user":          user,
		"access_token":  access,
		"refresh_token": refresh,
		"permissions":   rbac.Permissions(rbac.Role(user.Role)),
	})
}

// Refresh -> POST /auth/refresh
func (ctrl *AuthController) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	access, refresh, err := ctrl.authService.Refresh(req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			response.Error(c, http.StatusUnauthorized, "Invalid or expired refresh token", nil)
		case errors.Is(err, services.ErrUserDisabled):
			response.Error(c, http.StatusForbidden, "Your account is disabled", nil)
		default:
			response.InternalError(c, "Could not refresh token")
		}
		return
	}

	response.Success(c, "Token refreshed", gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

// Logout -> POST /auth/logout
func (ctrl *AuthController) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", response.ValidationErrors(err))
		return
	}

	if err := ctrl.authService.Logout(req.RefreshToken); err != nil {
		response.InternalError(c, "Logout failed")
		return
	}
	response.Success(c, "Logged out successfully", nil)
}

// Me -> GET /auth/me  (protected). Returns the current user + their permissions
// so the frontend can render the right menus.
func (ctrl *AuthController) Me(c *gin.Context) {
	user, err := ctrl.userService.Get(middleware.UserID(c))
	if err != nil {
		response.InternalError(c, "Could not load profile")
		return
	}
	response.Success(c, "Profile fetched", gin.H{
		"user":        user,
		"permissions": rbac.Permissions(rbac.Role(user.Role)),
	})
}
