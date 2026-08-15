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

	// 1. Missing Authorization header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing header, got %d", rec.Code)
	}

	// 2. Invalid Token
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid token, got %d", rec.Code)
	}

	// 3. Valid Token
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
