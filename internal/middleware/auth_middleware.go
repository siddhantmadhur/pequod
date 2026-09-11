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

// RequireAuth ensures that the incoming request contains a valid Bearer JWT header.
// Tokens in URL query parameters are rejected here to prevent credential leakage in proxy and server logs.
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Authorization token not provided in request header",
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Authorization token not provided in request header",
			})
		}

		user, err := m.authService.ValidateToken(c.Request().Context(), tokenStr)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Token not valid or session revoked",
				"error":   err.Error(),
			})
		}

		c.Set(UserContextKey, user)
		return next(c)
	}
}

// RequireStreamAuth provides authentication specifically for media playback (HLS master playlist, segments, direct play).
// It accepts Bearer headers and falls back to ?token= query parameter for external players (MPV, VLC, <video> elements).
func (m *AuthMiddleware) RequireStreamAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var tokenStr string
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else if qToken := c.QueryParam("token"); qToken != "" {
			tokenStr = qToken
		}

		if tokenStr == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Streaming authorization token not provided",
			})
		}

		user, err := m.authService.ValidateToken(c.Request().Context(), tokenStr)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Streaming token not valid or session revoked",
				"error":   err.Error(),
			})
		}

		c.Set(UserContextKey, user)
		return next(c)
	}
}

// RequireAdmin ensures that the authenticated user possesses administrator privileges (RoleAdmin == 0).
func (m *AuthMiddleware) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return m.RequireAuth(func(c echo.Context) error {
		user := CurrentUser(c)
		if user == nil || user.PermissionLevel != domain.RoleAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{
				"message": "Administrator privileges required",
			})
		}
		return next(c)
	})
}

// RequireAdminOrWizard allows unauthenticated access if the setup wizard is active, or requires admin otherwise.
func (m *AuthMiddleware) RequireAdminOrWizard(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !m.serverService.IsWizardCompleted() {
			return next(c)
		}
		return m.RequireAdmin(next)(c)
	}
}

// RequireAuthOrWizard allows unauthenticated access if setup wizard is active, or requires authentication otherwise.
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
