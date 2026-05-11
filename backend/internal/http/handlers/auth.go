package handlers

import (
	"net/http"

	"wa-platform/backend/internal/auth"
	"wa-platform/backend/internal/config"
	"wa-platform/backend/internal/store"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	Cfg   config.Config
	Store *store.Repository
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	// Query user from database
	user, err := h.Store.GetUserByUsername(c.Request().Context(), req.Username)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	// Verify password
	if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	token, err := auth.GenerateToken(h.Cfg.JWTSecret, user.Username, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to issue token"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]string{
			"username": user.Username,
			"role":     user.Role,
		},
	})
}
