package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/middleware"
	"github.com/taskflow/backend/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: name, email, and password (min 6 chars) are required")
		return
	}

	user, accessToken, refreshToken, err := h.authService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists):
			RespondWithError(c, http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrPasswordTooShort),
			errors.Is(err, service.ErrNameRequired),
			errors.Is(err, service.ErrEmailRequired):
			RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to register user")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: email and password are required")
		return
	}

	user, accessToken, refreshToken, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			RespondWithError(c, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to authenticate")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Refresh handles POST /api/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: refresh_token is required")
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		RespondWithError(c, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Me handles GET /api/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID := userIDVal.(uuid.UUID)
	user, err := h.authService.GetUserByID(userID)
	if err != nil || user == nil {
		RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
