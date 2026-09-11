package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/middleware"
)

type mockAuthService struct {
	validToken string
	user       *domain.User
}

func (m *mockAuthService) Login(ctx context.Context, username, password, device, deviceName, clientName, clientVersion string) (*domain.TokenPair, error) {
	return nil, nil
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshTokenString string) (*domain.TokenPair, error) {
	return nil, nil
}

func (m *mockAuthService) CreateUser(ctx context.Context, currentUser *domain.User, username, password string, permLevel int64, isWizardMode bool) error {
	return nil
}

func (m *mockAuthService) ValidateToken(tokenString string) (*domain.User, error) {
	if tokenString == m.validToken {
		return m.user, nil
	}
	return nil, domain.ErrUnauthorized
}

func (m *mockAuthService) Logout(ctx context.Context, sessionID string) error {
	return nil
}

type mockServerService struct {
	wizardCompleted bool
}

func (m *mockServerService) GetServerInfo() (*domain.ServerInformation, error) {
	return nil, nil
}

func (m *mockServerService) FinishWizard() error {
	m.wizardCompleted = true
	return nil
}

func (m *mockServerService) IsWizardCompleted() bool {
	return m.wizardCompleted
}

func TestRequireAuth(t *testing.T) {
	e := echo.New()
	authSvc := &mockAuthService{
		validToken: "valid-jwt-token",
		user: &domain.User{
			UID:             1,
			Username:        "testuser",
			PermissionLevel: 0,
		},
	}
	serverSvc := &mockServerService{}
	mw := middleware.NewAuthMiddleware(authSvc, serverSvc)

	handler := mw.RequireAuth(func(c echo.Context) error {
		user := middleware.CurrentUser(c)
		if user == nil {
			return c.String(http.StatusInternalServerError, "user missing in context")
		}
		return c.String(http.StatusOK, user.Username)
	})

	// 1. Missing Authorization header and query param
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing header, got %d", rec.Code)
	}

	// 2. Invalid Token in Header
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid token, got %d", rec.Code)
	}

	// 3. Valid Token in Header
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-jwt-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for valid token, got %d", rec.Code)
	}
	if rec.Body.String() != "testuser" {
		t.Errorf("expected body testuser, got %s", rec.Body.String())
	}

	// 4. SEC-05: Valid Token via Query Parameter (?token=...)
	req = httptest.NewRequest(http.MethodGet, "/test?token=valid-jwt-token", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for valid query param token, got %d", rec.Code)
	}
	if rec.Body.String() != "testuser" {
		t.Errorf("expected body testuser, got %s", rec.Body.String())
	}
}

func TestRequireAdmin(t *testing.T) {
	e := echo.New()
	adminUser := &domain.User{UID: 1, Username: "admin", PermissionLevel: 0}
	regularUser := &domain.User{UID: 2, Username: "regular", PermissionLevel: 1}

	authSvc := &mockAuthService{
		validToken: "valid-admin-token",
		user:       adminUser,
	}
	serverSvc := &mockServerService{}
	mw := middleware.NewAuthMiddleware(authSvc, serverSvc)

	adminHandler := mw.RequireAdmin(func(c echo.Context) error {
		return c.String(http.StatusOK, "admin-access-granted")
	})

	// 1. Admin user -> 200 OK
	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer valid-admin-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = adminHandler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for admin user, got %d", rec.Code)
	}

	// 2. Regular non-admin user -> 403 Forbidden (SEC-04)
	authSvc.user = regularUser
	req = httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer valid-admin-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = adminHandler(c)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-admin user, got %d", rec.Code)
	}
}

func TestRequireWizardActive(t *testing.T) {
	e := echo.New()
	authSvc := &mockAuthService{}
	serverSvc := &mockServerService{wizardCompleted: false}
	mw := middleware.NewAuthMiddleware(authSvc, serverSvc)

	handler := mw.RequireWizardActive(func(c echo.Context) error {
		return c.String(http.StatusOK, "wizard-ok")
	})

	// When wizard is active (not completed)
	req := httptest.NewRequest(http.MethodGet, "/wizard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when wizard is active, got %d", rec.Code)
	}

	// When wizard is completed
	serverSvc.wizardCompleted = true
	req = httptest.NewRequest(http.MethodGet, "/wizard", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 when wizard is completed, got %d", rec.Code)
	}
}

func TestRequireAdminOrWizard(t *testing.T) {
	e := echo.New()
	adminUser := &domain.User{UID: 1, Username: "admin", PermissionLevel: 0}
	regularUser := &domain.User{UID: 2, Username: "regular", PermissionLevel: 1}

	authSvc := &mockAuthService{
		validToken: "valid-token",
		user:       adminUser,
	}
	serverSvc := &mockServerService{wizardCompleted: false}
	mw := middleware.NewAuthMiddleware(authSvc, serverSvc)

	handler := mw.RequireAdminOrWizard(func(c echo.Context) error {
		return c.String(http.StatusOK, "admin-or-wizard-ok")
	})

	// 1. Wizard active allows unauthenticated access
	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 during wizard mode, got %d", rec.Code)
	}

	// 2. Wizard completed + Admin user -> 200 OK
	serverSvc.wizardCompleted = true
	req = httptest.NewRequest(http.MethodGet, "/setup", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for admin after wizard completed, got %d", rec.Code)
	}

	// 3. Wizard completed + Regular user -> 403 Forbidden
	authSvc.user = regularUser
	req = httptest.NewRequest(http.MethodGet, "/setup", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for regular user after wizard completed, got %d", rec.Code)
	}

	// 4. Wizard completed + Unauthenticated -> 401 Unauthorized
	req = httptest.NewRequest(http.MethodGet, "/setup", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated after wizard completed, got %d", rec.Code)
	}
}

func TestRequireAuthOrWizard(t *testing.T) {
	e := echo.New()
	authUser := &domain.User{UID: 2, Username: "regular", PermissionLevel: 1}
	authSvc := &mockAuthService{
		validToken: "valid-token",
		user:       authUser,
	}
	serverSvc := &mockServerService{wizardCompleted: false}
	mw := middleware.NewAuthMiddleware(authSvc, serverSvc)

	handler := mw.RequireAuthOrWizard(func(c echo.Context) error {
		return c.String(http.StatusOK, "auth-or-wizard-ok")
	})

	// 1. Wizard active allows unauthenticated access
	req := httptest.NewRequest(http.MethodGet, "/library", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when wizard is active, got %d", rec.Code)
	}

	// 2. Wizard completed + authenticated user -> 200 OK
	serverSvc.wizardCompleted = true
	req = httptest.NewRequest(http.MethodGet, "/library", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when authenticated after wizard, got %d", rec.Code)
	}

	// 3. Wizard completed + unauthenticated -> 401 Unauthorized
	req = httptest.NewRequest(http.MethodGet, "/library", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 when unauthenticated after wizard, got %d", rec.Code)
	}
}

