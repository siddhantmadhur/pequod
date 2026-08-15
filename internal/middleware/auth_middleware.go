package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/services"
)

const UserContextKey = "authenticated_user"

type AuthMiddleware struct {
	authService   services.AuthService
	serverService services.ServerService
}

func NewAuthMiddleware(authService services.AuthService, serverService services.ServerService) *AuthMiddleware {
	return &AuthMiddleware{
		authService:   authService,
		serverService: serverService,
	}
}

// CurrentUser extracts the authenticated user from the echo context.
func CurrentUser(c echo.Context) *domain.User {
	u, ok := c.Get(UserContextKey).(*domain.User)
	if !ok {
		return nil
	}
	return u
}

// RequireAuth ensures that the incoming request contains a valid Bearer JWT.
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Authorization token not provided",
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Authorization token not provided",
			})
		}

		user, err := m.authService.ValidateToken(tokenStr)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Token not valid",
				"error":   err.Error(),
			})
		}

		c.Set(UserContextKey, user)
		return next(c)
	}
}

// RequireAuthOrWizard allows unauthenticated access if setup wizard is active.
func (m *AuthMiddleware) RequireAuthOrWizard(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !m.serverService.IsWizardCompleted() {
			return next(c)
		}
		return m.RequireAuth(next)(c)
	}
}

// RequireWizardActive blocks access once the initial wizard has finished.
func (m *AuthMiddleware) RequireWizardActive(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if m.serverService.IsWizardCompleted() {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"msg": "Server is already setup and you no longer have access to this feature",
			})
		}
		return next(c)
	}
}
