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

// Login godoc
// @Summary  Log in and receive access + refresh tokens
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    body  body      LoginRequest  true  "Credentials"
// @Success  200   {object}  map[string]interface{}
// @Failure  401   {object}  map[string]interface{}
// @Router   /auth/login [post]
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

// Refresh godoc
// @Summary  Rotate the refresh token and get a new access token
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    body  body      RefreshRequest  true  "Refresh token"
// @Success  200   {object}  map[string]interface{}
// @Failure  401   {object}  map[string]interface{}
// @Router   /auth/refresh [post]
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

// Logout godoc
// @Summary  Revoke a refresh token (logout)
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    body  body      LogoutRequest  true  "Refresh token"
// @Success  200   {object}  map[string]interface{}
// @Router   /auth/logout [post]
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

// Me godoc
// @Summary  Get the current authenticated user + permissions
// @Tags     Auth
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  map[string]interface{}
// @Router   /auth/me [get]
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
