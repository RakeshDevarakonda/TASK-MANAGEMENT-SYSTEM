package handlers

import (
	"errors"
	"net/http"

	"task-system/internal/models"
	"task-system/internal/repository"
	"task-system/internal/service"
	"task-system/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{service: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	if req.Password != req.ConfirmPassword {
		utils.SendError(c, http.StatusBadRequest, "passwords do not match", nil)
		return
	}

	user, err := h.service.Register(req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			utils.SendError(c, http.StatusConflict, "email already exists", nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "failed to create user record", nil)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "User registered successfully", user)
}

func (h *AuthHandler) Signin(c *gin.Context) {
	var req models.SigninRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	user, err := h.service.Signin(req.Email, req.Password)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}

	// Create session, hashing the refresh token, and caching session/user state in Redis
	accessToken, refreshToken, err := h.service.CreateSession(user)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to generate session", nil)
		return
	}

	// Write refresh token and access token to secure, HTTP-Only cookies (30 days refresh TTL)
	c.SetCookie("refreshToken", refreshToken, 30*24*3600, "/", "", true, true)
	c.SetCookie("accessToken", accessToken, 15*60, "/", "", true, true)

	utils.SendSuccess(c, http.StatusOK, "User signed in successfully", gin.H{
		"User": user,
		"accessToken": accessToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	// Read Refresh Token from HTTP-Only cookie
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, "missing refresh token", nil)
		return
	}

	// Generate new rotated access and refresh tokens
	newAccessToken, newRefreshToken, err := h.service.RefreshSession(refreshToken)
	if err != nil {
		// Clear cookies immediately on session expiration or invalidation
		c.SetCookie("refreshToken", "", -1, "/", "", true, true)
		c.SetCookie("accessToken", "", -1, "/", "", true, true)
		utils.SendError(c, http.StatusUnauthorized, "session expired or invalid", nil)
		return
	}

	// Write rotated tokens to secure, HTTP-Only cookies (extend refresh TTL to 30 days)
	c.SetCookie("refreshToken", newRefreshToken, 30*24*3600, "/", "", true, true)
	c.SetCookie("accessToken", newAccessToken, 15*60, "/", "", true, true)

	utils.SendSuccess(c, http.StatusOK, "token refreshed successfully", gin.H{"accessToken": newAccessToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie("refreshToken")

	// Get authenticated user ID from context
	var userID string
	if currentUser, exists := c.Get("currentUser"); exists {
		if u, ok := currentUser.(*models.User); ok {
			userID = u.ID
		}
	}

	// Invalidate session from Postgres and Redis if we have session info
	if refreshToken != "" && userID != "" {
		_ = h.service.InvalidateSession(refreshToken, userID)
	}

	// Clear HTTP-Only cookies immediately
	c.SetCookie("accessToken", "", -1, "/", "", true, true)
	c.SetCookie("refreshToken", "", -1, "/", "", true, true)

	utils.SendSuccess(c, http.StatusOK, "logged out successfully", nil)
}
