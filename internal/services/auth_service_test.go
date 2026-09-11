package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/services"
	"golang.org/x/crypto/bcrypt"
)

// --- Mocks & Test Fixtures ---

type mockUserRepo struct {
	profiles map[string]*domain.Profile
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		profiles: make(map[string]*domain.Profile),
	}
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

type mockSessionRepo struct {
	sessions map[string]*domain.Session
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[string]*domain.Session),
	}
}

func (m *mockSessionRepo) Create(ctx context.Context, s *domain.Session) error {
	m.sessions[s.ID] = s
	return nil
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return s, nil
}

func (m *mockSessionRepo) Delete(ctx context.Context, id string) error {
	delete(m.sessions, id)
	return nil
}

func (m *mockSessionRepo) DeleteByUserID(ctx context.Context, userID int64) error {
	for id, s := range m.sessions {
		if s.UserID == userID {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockSessionRepo) DeleteExpired(ctx context.Context, before time.Time) error {
	for id, s := range m.sessions {
		if s.RefreshExpiresAt.Before(before) {
			delete(m.sessions, id)
		}
	}
	return nil
}

type testAuthSetup struct {
	userRepo    *mockUserRepo
	sessionRepo *mockSessionRepo
	authService services.AuthService
	secretKey   string
}

func newTestAuthSetup() *testAuthSetup {
	secret := "test-secret-key-32-chars-long-abc"
	userRepo := newMockUserRepo()
	sessionRepo := newMockSessionRepo()

	// Seed default admin user: username="admin", password="password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	userRepo.profiles["admin"] = &domain.Profile{
		ID:       1,
		Username: "admin",
		Password: hashedPassword,
		Type:     0, // Admin
	}

	authService := services.NewAuthService(userRepo, sessionRepo, secret)

	return &testAuthSetup{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		authService: authService,
		secretKey:   secret,
	}
}

// --- Unit Tests ---

func TestAuthService_Login(t *testing.T) {
	setup := newTestAuthSetup()
	ctx := context.Background()

	t.Run("successful login generates tokens and records session", func(t *testing.T) {
		tokens, err := setup.authService.Login(ctx, "admin", "password123", "desktop", "MacBook Pro", "Pequod Client", "1.0")
		if err != nil {
			t.Fatalf("unexpected login error: %v", err)
		}
		if tokens == nil || tokens.AccessToken == "" || tokens.RefreshToken == "" {
			t.Fatal("expected non-empty access and refresh tokens")
		}

		// Verify session was persisted in repository
		if len(setup.sessionRepo.sessions) != 1 {
			t.Fatalf("expected 1 session in repository, found %d", len(setup.sessionRepo.sessions))
		}
	})

	t.Run("login with non-existent user fails", func(t *testing.T) {
		_, err := setup.authService.Login(ctx, "ghost_user", "password123", "", "", "", "")
		if err != domain.ErrInvalidCredentials {
			t.Fatalf("expected ErrInvalidCredentials for non-existent user, got: %v", err)
		}
	})

	t.Run("login with incorrect password fails", func(t *testing.T) {
		_, err := setup.authService.Login(ctx, "admin", "wrong_password_999", "", "", "", "")
		if err != domain.ErrInvalidCredentials {
			t.Fatalf("expected ErrInvalidCredentials for wrong password, got: %v", err)
		}
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	setup := newTestAuthSetup()
	ctx := context.Background()

	tokens, err := setup.authService.Login(ctx, "admin", "password123", "device1", "Mac", "Client", "1.0")
	if err != nil {
		t.Fatalf("setup login failed: %v", err)
	}

	t.Run("valid access token returns authenticated user with session ID", func(t *testing.T) {
		user, err := setup.authService.ValidateToken(ctx, tokens.AccessToken)
		if err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
		if user == nil {
			t.Fatal("expected non-nil user")
		}
		if user.Username != "admin" || user.UID != 1 {
			t.Fatalf("expected user admin with UID 1, got username=%s, UID=%d", user.Username, user.UID)
		}
		if user.SessionID == "" {
			t.Fatal("expected user to contain non-empty SessionID")
		}
	})

	t.Run("SEC-02: refresh token rejected when presented for API validation", func(t *testing.T) {
		_, err := setup.authService.ValidateToken(ctx, tokens.RefreshToken)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized when validating refresh token as access token, got: %v", err)
		}
	})

	t.Run("token signed with different secret is rejected", func(t *testing.T) {
		otherService := services.NewAuthService(setup.userRepo, setup.sessionRepo, "completely-different-secret-key")
		_, err := otherService.ValidateToken(ctx, tokens.AccessToken)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized for token signed with different secret, got: %v", err)
		}
	})

	t.Run("tampered or malformed token string is rejected", func(t *testing.T) {
		testCases := []string{
			"",
			"not.a.token",
			tokens.AccessToken + "tampered",
		}
		for _, tc := range testCases {
			_, err := setup.authService.ValidateToken(ctx, tc)
			if err != domain.ErrUnauthorized {
				t.Fatalf("expected ErrUnauthorized for malformed token %q, got: %v", tc, err)
			}
		}
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"token_type": "access",
			"username":   "admin",
			"uid":        "1",
			"exp":        time.Now().Add(-time.Hour).Unix(),
		})
		tokenStr, _ := expiredToken.SignedString([]byte(setup.secretKey))

		_, err := setup.authService.ValidateToken(ctx, tokenStr)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized for expired token, got: %v", err)
		}
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	setup := newTestAuthSetup()
	ctx := context.Background()

	tokens, err := setup.authService.Login(ctx, "admin", "password123", "device1", "Mac", "Client", "1.0")
	if err != nil {
		t.Fatalf("setup login failed: %v", err)
	}

	t.Run("SEC-03: valid refresh token rotates session and issues new token pair", func(t *testing.T) {
		refreshed, err := setup.authService.RefreshToken(ctx, tokens.RefreshToken)
		if err != nil {
			t.Fatalf("unexpected refresh error: %v", err)
		}
		if refreshed == nil || refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
			t.Fatal("expected new token pair from refresh")
		}

		// Replay protection: the old refresh token must be rotated out and cannot be used again
		_, err = setup.authService.RefreshToken(ctx, tokens.RefreshToken)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized when reusing rotated refresh token, got: %v", err)
		}

		// The new refresh token should be usable
		newRefreshed, err := setup.authService.RefreshToken(ctx, refreshed.RefreshToken)
		if err != nil {
			t.Fatalf("expected new refresh token to be valid, got: %v", err)
		}
		if newRefreshed == nil || newRefreshed.AccessToken == "" {
			t.Fatal("expected valid tokens from second refresh")
		}
	})

	t.Run("SEC-02: access token cannot be used at refresh endpoint", func(t *testing.T) {
		_, err := setup.authService.RefreshToken(ctx, tokens.AccessToken)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized when using access token as refresh token, got: %v", err)
		}
	})

	t.Run("SEC-03: refresh token with revoked/deleted session in DB is rejected", func(t *testing.T) {
		// Log in another device
		devTokens, err := setup.authService.Login(ctx, "admin", "password123", "mobile", "iPhone", "App", "1.0")
		if err != nil {
			t.Fatalf("device login failed: %v", err)
		}

		// Validate to get session ID and delete session from DB (simulating remote logout or admin revocation)
		u, err := setup.authService.ValidateToken(ctx, devTokens.AccessToken)
		if err != nil {
			t.Fatalf("token validation failed: %v", err)
		}
		_ = setup.sessionRepo.Delete(ctx, u.SessionID)

		// Refresh must now fail because session does not exist in DB
		_, err = setup.authService.RefreshToken(ctx, devTokens.RefreshToken)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized when session has been revoked in DB, got: %v", err)
		}
	})

	t.Run("SEC-03: expired session is rejected and purged", func(t *testing.T) {
		expTokens, err := setup.authService.Login(ctx, "admin", "password123", "tv", "AppleTV", "App", "1.0")
		if err != nil {
			t.Fatalf("tv login failed: %v", err)
		}

		u, _ := setup.authService.ValidateToken(ctx, expTokens.AccessToken)
		// Manually backdate the session's RefreshExpiresAt in repo
		if sess, ok := setup.sessionRepo.sessions[u.SessionID]; ok {
			sess.RefreshExpiresAt = time.Now().Add(-24 * time.Hour)
		}

		_, err = setup.authService.RefreshToken(ctx, expTokens.RefreshToken)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized for expired session, got: %v", err)
		}

		// Ensure expired session was removed from repo
		if _, exists := setup.sessionRepo.sessions[u.SessionID]; exists {
			t.Fatal("expected expired session to be deleted from repo")
		}
	})
}

func TestAuthService_Logout(t *testing.T) {
	setup := newTestAuthSetup()
	ctx := context.Background()

	tokens, err := setup.authService.Login(ctx, "admin", "password123", "device", "Mac", "Client", "1.0")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	user, err := setup.authService.ValidateToken(ctx, tokens.AccessToken)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}

	t.Run("logging out deletes the session from the repository and immediately revokes access token", func(t *testing.T) {
		err := setup.authService.Logout(ctx, user.SessionID)
		if err != nil {
			t.Fatalf("unexpected logout error: %v", err)
		}

		// Session should no longer exist in DB
		_, err = setup.sessionRepo.GetByID(ctx, user.SessionID)
		if err != domain.ErrNotFound {
			t.Fatalf("expected ErrNotFound for deleted session, got: %v", err)
		}

		// Reviewer check: ValidateToken must immediately reject the access token post-logout
		_, err = setup.authService.ValidateToken(ctx, tokens.AccessToken)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized when validating access token post-logout, got: %v", err)
		}

		// Refresh with that session must also fail
		_, err = setup.authService.RefreshToken(ctx, tokens.RefreshToken)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized after logging out, got: %v", err)
		}
	})

	t.Run("logout with empty sessionID is safe no-op", func(t *testing.T) {
		err := setup.authService.Logout(ctx, "")
		if err != nil {
			t.Fatalf("expected nil error for empty sessionID logout, got: %v", err)
		}
	})
}

func TestAuthService_CreateUser_Validation(t *testing.T) {
	setup := newTestAuthSetup()
	ctx := context.Background()
	adminUser := &domain.User{UID: 1, Username: "admin", PermissionLevel: 0}

	testCases := []struct {
		name        string
		username    string
		password    string
		expectedErr error
	}{
		{
			name:        "SEC-06: password under 8 characters rejected",
			username:    "validuser",
			password:    "1234567",
			expectedErr: domain.ErrPasswordTooShort,
		},
		{
			name:        "SEC-06: password over 72 characters rejected (bcrypt safeguard)",
			username:    "validuser",
			password:    "1234567890123456789012345678901234567890123456789012345678901234567890123", // 73 chars
			expectedErr: domain.ErrPasswordTooLong,
		},
		{
			name:        "SEC-06: username under 3 characters rejected",
			username:    "ab",
			password:    "password123",
			expectedErr: domain.ErrInvalidUsername,
		},
		{
			name:        "SEC-06: username over 64 characters rejected",
			username:    "a12345678901234567890123456789012345678901234567890123456789012345", // 65 chars
			password:    "password123",
			expectedErr: domain.ErrInvalidUsername,
		},
		{
			name:        "SEC-06: username with spaces rejected",
			username:    "john doe",
			password:    "password123",
			expectedErr: domain.ErrInvalidUsername,
		},
		{
			name:        "SEC-06: username with special characters rejected",
			username:    "john@domain!#",
			password:    "password123",
			expectedErr: domain.ErrInvalidUsername,
		},
		{
			name:        "duplicate username rejected",
			username:    "admin",
			password:    "password123",
			expectedErr: domain.ErrUserAlreadyExists,
		},
		{
			name:        "valid username and password succeeds",
			username:    "valid.user_1-ok",
			password:    "securePass123!",
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := setup.authService.CreateUser(ctx, adminUser, tc.username, tc.password, 1, false)
			if err != tc.expectedErr {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestAuthService_CreateUser_WizardMode(t *testing.T) {
	ctx := context.Background()

	t.Run("first user creation in wizard mode creates admin user", func(t *testing.T) {
		emptyUserRepo := newMockUserRepo()
		emptySessionRepo := newMockSessionRepo()
		authSvc := services.NewAuthService(emptyUserRepo, emptySessionRepo, "secret-key-1234567890")

		err := authSvc.CreateUser(ctx, nil, "first_admin", "password123", 0, true)
		if err != nil {
			t.Fatalf("unexpected error creating first user in wizard mode: %v", err)
		}

		admin, err := emptyUserRepo.GetByUsername(ctx, "first_admin")
		if err != nil {
			t.Fatalf("failed to retrieve created first user: %v", err)
		}
		if admin.Type != 0 {
			t.Fatalf("expected first user to have admin permission 0, got %d", admin.Type)
		}
	})

	t.Run("SEC-01: second user or takeover attempt during wizard mode is blocked", func(t *testing.T) {
		setup := newTestAuthSetup() // already has admin seeded (count == 1)

		// Attacker attempts to call wizard user creation while admin exists
		err := setup.authService.CreateUser(ctx, nil, "attacker", "password123", 0, true)
		if err != domain.ErrUserAlreadyExists {
			t.Fatalf("expected ErrUserAlreadyExists when wizard mode called with existing admin, got: %v", err)
		}

		// Ensure the attacker account was not created
		_, err = setup.userRepo.GetByUsername(ctx, "attacker")
		if err != domain.ErrNotFound {
			t.Fatal("attacker user should not have been created")
		}
	})
}

func TestAuthService_CreateUser_Permissions(t *testing.T) {
	setup := newTestAuthSetup()
	ctx := context.Background()

	adminUser := &domain.User{UID: 1, Username: "admin", PermissionLevel: 0}
	regularUser := &domain.User{UID: 2, Username: "regular", PermissionLevel: 1}

	t.Run("unauthenticated user cannot create users outside wizard mode", func(t *testing.T) {
		err := setup.authService.CreateUser(ctx, nil, "new_member", "password123", 1, false)
		if err != domain.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized when currentUser is nil, got: %v", err)
		}
	})

	t.Run("SEC-04: non-admin user cannot create other users", func(t *testing.T) {
		err := setup.authService.CreateUser(ctx, regularUser, "new_member", "password123", 1, false)
		if err != domain.ErrForbidden {
			t.Fatalf("expected ErrForbidden for non-admin user, got: %v", err)
		}
	})

	t.Run("SEC-04: admin user can create new users", func(t *testing.T) {
		err := setup.authService.CreateUser(ctx, adminUser, "new_member", "password123", 1, false)
		if err != nil {
			t.Fatalf("expected admin to create user successfully, got: %v", err)
		}

		created, err := setup.userRepo.GetByUsername(ctx, "new_member")
		if err != nil || created == nil {
			t.Fatal("expected user to be created in repository")
		}
		if created.Type != 1 {
			t.Fatalf("expected permission level 1, got %d", created.Type)
		}
	})
}
