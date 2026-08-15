package services_test

import (
	"context"
	"testing"

	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/services"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	profiles map[string]*domain.Profile
}

func (m *mockUserRepo) Create(ctx context.Context, username string, password []byte, permLevel int64) error {
	m.profiles[username] = &domain.Profile{
		ID:       int64(len(m.profiles) + 1),
		Username: username,
		Password: password,
		Type:     permLevel,
	}
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, id int64, username string, password []byte) error {
	for _, p := range m.profiles {
		if p.ID == id {
			p.Username = username
			p.Password = password
			return nil
		}
	}
	return domain.ErrNotFound
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*domain.Profile, error) {
	for _, p := range m.profiles {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	p, ok := m.profiles[username]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockUserRepo) GetAdmin(ctx context.Context) (*domain.Profile, error) {
	for _, p := range m.profiles {
		if p.Type == 0 {
			return p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserRepo) GetAll(ctx context.Context) ([]domain.Profile, error) {
	var list []domain.Profile
	for _, p := range m.profiles {
		list = append(list, *p)
	}
	return list, nil
}

func (m *mockUserRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.profiles)), nil
}

type mockSessionRepo struct{}

func (m *mockSessionRepo) Create(ctx context.Context, s *domain.Session) error {
	return nil
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	return nil, domain.ErrNotFound
}

func TestAuthService_LoginAndValidation(t *testing.T) {
	secret := "test-secret-key-12345"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)

	userRepo := &mockUserRepo{
		profiles: map[string]*domain.Profile{
			"admin": {
				ID:       1,
				Username: "admin",
				Password: hashedPassword,
				Type:     0,
			},
		},
	}
	sessionRepo := &mockSessionRepo{}
	authService := services.NewAuthService(userRepo, sessionRepo, secret)

	// 1. Successful Login
	tokens, err := authService.Login(context.Background(), "admin", "password123", "mac", "MacBook", "Web", "1.0")
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}

	// 2. Validate Access Token
	user, err := authService.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("unexpected validate token error: %v", err)
	}
	if user.Username != "admin" || user.UID != 1 {
		t.Fatalf("expected user admin with ID 1, got %s / %d", user.Username, user.UID)
	}

	// 3. Failed Login (Wrong Password)
	_, err = authService.Login(context.Background(), "admin", "wrongpassword", "", "", "", "")
	if err == nil {
		t.Fatal("expected error on wrong password, got nil")
	}

	// 4. Refresh Token
	refreshed, err := authService.RefreshToken(context.Background(), tokens.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected refresh error: %v", err)
	}
	if refreshed.AccessToken == "" {
		t.Fatal("expected refreshed access token")
	}

	// 5. Test Permission Check in CreateUser
	regularUser := &domain.User{
		UID:             2,
		Username:        "regular",
		PermissionLevel: 1, // Non-admin
	}
	adminUser := &domain.User{
		UID:             1,
		Username:        "admin",
		PermissionLevel: 0, // Admin
	}

	err = authService.CreateUser(context.Background(), regularUser, "newuser", "pass123", 1, false)
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden for non-admin user creating user, got: %v", err)
	}

	err = authService.CreateUser(context.Background(), adminUser, "newuser", "pass123", 1, false)
	if err != nil {
		t.Fatalf("expected admin user to create user successfully, got: %v", err)
	}

	// 6. Test Invalid / Tampered Token
	wrongService := services.NewAuthService(userRepo, sessionRepo, "different-secret-key")
	_, err = wrongService.ValidateToken(tokens.AccessToken)
	if err != domain.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized for token with wrong secret, got: %v", err)
	}
}
