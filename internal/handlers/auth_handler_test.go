package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/handlers"
	"github.com/siddhantmadhur/pequod/internal/middleware"
)

// --- Mock AuthService for Handler Tests ---

type mockAuthServiceForHandler struct {
	loginFunc        func(ctx context.Context, username, password string) (*domain.TokenPair, error)
	refreshTokenFunc func(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	createUserFunc   func(ctx context.Context, currentUser *domain.User, username, password string, permLevel int64, isWizard bool) error
	validateFunc     func(ctx context.Context, tokenString string) (*domain.User, error)
	logoutFunc       func(ctx context.Context, sessionID string) error
}

func (m *mockAuthServiceForHandler) Login(ctx context.Context, username, password, device, deviceName, clientName, clientVersion string) (*domain.TokenPair, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, username, password)
	}
	return &domain.TokenPair{AccessToken: "mock-access-token", RefreshToken: "mock-refresh-token"}, nil
}

func (m *mockAuthServiceForHandler) RefreshToken(ctx context.Context, refreshTokenString string) (*domain.TokenPair, error) {
	if m.refreshTokenFunc != nil {
		return m.refreshTokenFunc(ctx, refreshTokenString)
	}
	return &domain.TokenPair{AccessToken: "new-access-token", RefreshToken: "new-refresh-token"}, nil
}

func (m *mockAuthServiceForHandler) CreateUser(ctx context.Context, currentUser *domain.User, username, password string, permLevel int64, isWizardMode bool) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, currentUser, username, password, permLevel, isWizardMode)
	}
	return nil
}

func (m *mockAuthServiceForHandler) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, tokenString)
	}
	return &domain.User{UID: 1, Username: "admin", PermissionLevel: domain.RoleAdmin}, nil
}

func (m *mockAuthServiceForHandler) Logout(ctx context.Context, sessionID string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, sessionID)
	}
	return nil
}

// --- Handler Tests ---

func TestAuthHandler_Login(t *testing.T) {
	e := echo.New()
	mockAuth := &mockAuthServiceForHandler{}
	mockServer := &MockServerService{FinishedWizard: true}
	handler := handlers.NewAuthHandler(mockAuth, mockServer)

	t.Run("successful login returns 200 and token pair", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": "admin",
			"password": "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Login(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode JSON response: %v", err)
		}
		data, ok := res["data"].(map[string]interface{})
		if !ok || data["access_token"] != "mock-access-token" {
			t.Fatalf("expected data.access_token in response, got %v", res)
		}
	})

	t.Run("invalid credentials returns 400 Bad Request", func(t *testing.T) {
		mockAuth.loginFunc = func(ctx context.Context, username, password string) (*domain.TokenPair, error) {
			return nil, domain.ErrInvalidCredentials
		}
		defer func() { mockAuth.loginFunc = nil }()

		body, _ := json.Marshal(map[string]string{
			"username": "admin",
			"password": "wrongpassword",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.Login(c)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for invalid credentials, got %d", rec.Code)
		}
	})

	t.Run("malformed json payload returns 400 Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("{invalid-json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.Login(c)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for bad json, got %d", rec.Code)
		}
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	e := echo.New()
	mockAuth := &mockAuthServiceForHandler{}
	mockServer := &MockServerService{FinishedWizard: true}
	handler := handlers.NewAuthHandler(mockAuth, mockServer)

	t.Run("valid Bearer refresh token returns 200 and new tokens", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer valid-refresh-token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.RefreshToken(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var res map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		data := res["data"].(map[string]interface{})
		if data["access_token"] != "new-access-token" {
			t.Fatalf("expected new-access-token, got %v", data)
		}
	})

	t.Run("missing authorization header returns 401 Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.RefreshToken(c)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 for missing header, got %d", rec.Code)
		}
	})

	t.Run("service rejection returns 401 Unauthorized", func(t *testing.T) {
		mockAuth.refreshTokenFunc = func(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
			return nil, domain.ErrUnauthorized
		}
		defer func() { mockAuth.refreshTokenFunc = nil }()

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer expired-token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.RefreshToken(c)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})
}

func TestAuthHandler_CreateUser(t *testing.T) {
	e := echo.New()
	mockAuth := &mockAuthServiceForHandler{}
	mockServer := &MockServerService{FinishedWizard: true}
	handler := handlers.NewAuthHandler(mockAuth, mockServer)

	t.Run("successful user creation returns 201 Created", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"username":       "newuser",
			"password":       "password123",
			"permission_int": 1,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/create/user", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Set authenticated admin user in context
		c.Set(middleware.UserContextKey, &domain.User{UID: 1, Username: "admin", PermissionLevel: 0})

		err := handler.CreateUser(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d", rec.Code)
		}
	})

	t.Run("password too short returns 400 Bad Request", func(t *testing.T) {
		mockAuth.createUserFunc = func(ctx context.Context, currentUser *domain.User, username, password string, permLevel int64, isWizard bool) error {
			return domain.ErrPasswordTooShort
		}
		defer func() { mockAuth.createUserFunc = nil }()

		body, _ := json.Marshal(map[string]interface{}{
			"username": "newuser",
			"password": "123",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/create/user", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middleware.UserContextKey, &domain.User{UID: 1, Username: "admin", PermissionLevel: 0})

		_ = handler.CreateUser(c)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for short password, got %d", rec.Code)
		}
	})

	t.Run("duplicate user returns 409 Conflict", func(t *testing.T) {
		mockAuth.createUserFunc = func(ctx context.Context, currentUser *domain.User, username, password string, permLevel int64, isWizard bool) error {
			return domain.ErrUserAlreadyExists
		}
		defer func() { mockAuth.createUserFunc = nil }()

		body, _ := json.Marshal(map[string]interface{}{
			"username": "existinguser",
			"password": "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/create/user", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middleware.UserContextKey, &domain.User{UID: 1, Username: "admin", PermissionLevel: 0})

		_ = handler.CreateUser(c)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status 409 for duplicate user, got %d", rec.Code)
		}
	})

	t.Run("non-admin caller returns 403 Forbidden", func(t *testing.T) {
		mockAuth.createUserFunc = func(ctx context.Context, currentUser *domain.User, username, password string, permLevel int64, isWizard bool) error {
			return domain.ErrForbidden
		}
		defer func() { mockAuth.createUserFunc = nil }()

		body, _ := json.Marshal(map[string]interface{}{
			"username": "newuser",
			"password": "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/create/user", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middleware.UserContextKey, &domain.User{UID: 2, Username: "regular", PermissionLevel: 1})

		_ = handler.CreateUser(c)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403 for non-admin, got %d", rec.Code)
		}
	})
}

func TestAuthHandler_GetUser(t *testing.T) {
	e := echo.New()
	mockAuth := &mockAuthServiceForHandler{}
	mockServer := &MockServerService{}
	handler := handlers.NewAuthHandler(mockAuth, mockServer)

	t.Run("authenticated user returns 200 and user profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/user", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middleware.UserContextKey, &domain.User{
			UID:      1,
			Username: "admin",
		})

		err := handler.GetUser(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var u domain.User
		_ = json.Unmarshal(rec.Body.Bytes(), &u)
		if u.Username != "admin" || u.UID != 1 {
			t.Fatalf("expected user admin with UID 1, got %+v", u)
		}
	})

	t.Run("unauthenticated context returns 401 Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/user", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.GetUser(c)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	e := echo.New()
	loggedOutSession := ""
	mockAuth := &mockAuthServiceForHandler{
		logoutFunc: func(ctx context.Context, sessionID string) error {
			loggedOutSession = sessionID
			return nil
		},
	}
	mockServer := &MockServerService{}
	handler := handlers.NewAuthHandler(mockAuth, mockServer)

	t.Run("authenticated user logout invokes service and returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(middleware.UserContextKey, &domain.User{
			UID:       1,
			Username:  "admin",
			SessionID: "test-session-uuid-1234",
		})

		err := handler.Logout(c)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if loggedOutSession != "test-session-uuid-1234" {
			t.Fatalf("expected session test-session-uuid-1234 to be revoked, got %q", loggedOutSession)
		}

		var res map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res["message"] != "Successfully logged out" {
			t.Fatalf("expected logout success message, got %v", res)
		}
	})
}
