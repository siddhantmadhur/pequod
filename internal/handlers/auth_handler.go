package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/middleware"
	"github.com/siddhantmadhur/pequod/internal/services"
)

type AuthHandler struct {
	authService   services.AuthService
	serverService services.ServerService
}

func NewAuthHandler(authService services.AuthService, serverService services.ServerService) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		serverService: serverService,
	}
}

type LoginRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Device        string `json:"device"`
	DeviceName    string `json:"device_name"`
	ClientName    string `json:"client_name"`
	ClientVersion string `json:"client_version"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return HandleError(c, domain.ErrInvalidInput)
	}

	tokens, err := h.authService.Login(c.Request().Context(), req.Username, req.Password, req.Device, req.DeviceName, req.ClientName, req.ClientVersion)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"data": map[string]string{
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
		},
	})
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return HandleError(c, domain.ErrUnauthorized)
	}

	rawRefreshToken := strings.TrimPrefix(authHeader, "Bearer ")
	tokens, err := h.authService.RefreshToken(c.Request().Context(), rawRefreshToken)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Refresh successful",
		"data": map[string]string{
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
		},
	})
}

type CreateUserRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	PermissionInt int64  `json:"permission_int"`
}

func (h *AuthHandler) CreateUser(c echo.Context) error {
	currentUser := middleware.CurrentUser(c)
	isWizardMode := !h.serverService.IsWizardCompleted()

	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return HandleError(c, domain.ErrInvalidInput)
	}

	err := h.authService.CreateUser(c.Request().Context(), currentUser, req.Username, req.Password, req.PermissionInt, isWizardMode)
	if err != nil {
		return HandleError(c, err)
	}

	return c.NoContent(http.StatusCreated)
}

func (h *AuthHandler) GetUser(c echo.Context) error {
	currentUser := middleware.CurrentUser(c)
	if currentUser == nil {
		return HandleError(c, domain.ErrUnauthorized)
	}
	return c.JSON(http.StatusOK, currentUser)
}
