package services

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

type AuthService interface {
	Login(ctx context.Context, username, password, device, deviceName, clientName, clientVersion string) (*domain.TokenPair, error)
	RefreshToken(ctx context.Context, refreshTokenString string) (*domain.TokenPair, error)
	CreateUser(ctx context.Context, currentUser *domain.User, username, password string, permLevel int64, isWizardMode bool) error
	ValidateToken(tokenString string) (*domain.User, error)
	Logout(ctx context.Context, sessionID string) error
}

type authService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	secretKey   string
}

func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, secretKey string) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		secretKey:   secretKey,
	}
}

func (s *authService) Login(ctx context.Context, username, password, device, deviceName, clientName, clientVersion string) (*domain.TokenPair, error) {
	profile, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(profile.Password, []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	sessionID := uuid.NewString()
	tokens, err := s.generateTokens(profile, sessionID)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:               sessionID,
		UserID:           profile.ID,
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		CreatedAt:        time.Now(),
		AccessExpiresAt:  time.Now().Add(time.Minute * 20),
		RefreshExpiresAt: time.Now().Add(time.Hour * 300),
		Device:           device,
		DeviceName:       deviceName,
		ClientName:       clientName,
		ClientVersion:    clientVersion,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to persist user session: %w", err)
	}

	return tokens, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshTokenString string) (*domain.TokenPair, error) {
	token, err := jwt.Parse(refreshTokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	// SEC-02: Ensure this token is strictly a refresh token, NOT an access token
	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, domain.ErrUnauthorized
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return nil, domain.ErrUnauthorized
	}

	sid, _ := claims["sid"].(string)
	if sid == "" {
		return nil, domain.ErrUnauthorized
	}

	// SEC-03: Verify that the session is active and not revoked in the database
	activeSession, err := s.sessionRepo.GetByID(ctx, sid)
	if err != nil || activeSession == nil {
		return nil, domain.ErrUnauthorized
	}

	if time.Now().After(activeSession.RefreshExpiresAt) {
		_ = s.sessionRepo.Delete(ctx, sid)
		return nil, domain.ErrUnauthorized
	}

	profile, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Session rotation: revoke previous session and issue new session pair
	_ = s.sessionRepo.Delete(ctx, sid)

	newSessionID := uuid.NewString()
	tokens, err := s.generateTokens(profile, newSessionID)
	if err != nil {
		return nil, err
	}

	newSession := &domain.Session{
		ID:               newSessionID,
		UserID:           profile.ID,
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		CreatedAt:        time.Now(),
		AccessExpiresAt:  time.Now().Add(time.Minute * 20),
		RefreshExpiresAt: time.Now().Add(time.Hour * 300),
		Device:           activeSession.Device,
		DeviceName:       activeSession.DeviceName,
		ClientName:       activeSession.ClientName,
		ClientVersion:    activeSession.ClientVersion,
	}

	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, fmt.Errorf("failed to persist refreshed session: %w", err)
	}

	return tokens, nil
}

func (s *authService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return s.sessionRepo.Delete(ctx, sessionID)
}

func (s *authService) CreateUser(ctx context.Context, currentUser *domain.User, username, password string, permLevel int64, isWizardMode bool) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 64 || !usernameRegex.MatchString(username) {
		return domain.ErrInvalidUsername
	}

	// SEC-06: Validate password length bounds (bcrypt 72-byte max safeguard)
	if len(password) < 8 {
		return domain.ErrPasswordTooShort
	}
	if len(password) > 72 {
		return domain.ErrPasswordTooLong
	}

	// Check if username is already taken
	existing, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil && existing != nil {
		return domain.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	// SEC-01: In wizard mode, allow creating the first admin only if 0 users exist.
	// Overwriting an existing admin without credentials is strictly forbidden.
	if isWizardMode {
		count, err := s.userRepo.Count(ctx)
		if err != nil {
			return err
		}
		if count > 0 {
			return domain.ErrUserAlreadyExists
		}
		return s.userRepo.Create(ctx, username, hashedPassword, 0)
	}

	if currentUser == nil {
		return domain.ErrUnauthorized
	}
	if currentUser.PermissionLevel != 0 {
		return domain.ErrForbidden
	}

	return s.userRepo.Create(ctx, username, hashedPassword, permLevel)
}

func (s *authService) ValidateToken(tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	// SEC-02: Enforce token type to prevent using refresh tokens for API calls
	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "access" {
		return nil, domain.ErrUnauthorized
	}

	username, _ := claims["username"].(string)
	uidStr, _ := claims["uid"].(string)
	sid, _ := claims["sid"].(string)
	uid, _ := strconv.ParseInt(uidStr, 10, 64)

	var permLevel int
	if permVal, ok := claims["perm"].(float64); ok {
		permLevel = int(permVal)
	}

	return &domain.User{
		UID:             uid,
		Username:        username,
		AccessToken:     tokenString,
		RefreshToken:    "",
		PermissionLevel: permLevel,
		SessionID:       sid,
	}, nil
}

func (s *authService) generateTokens(profile *domain.Profile, sessionID string) (*domain.TokenPair, error) {
	now := time.Now()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token_type": "access",
		"username":   profile.Username,
		"uid":        fmt.Sprint(profile.ID),
		"perm":       profile.Type,
		"sid":        sessionID,
		"iat":        now.Unix(),
		"exp":        now.Add(time.Minute * 20).Unix(),
	})
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token_type": "refresh",
		"username":   profile.Username,
		"uid":        fmt.Sprint(profile.ID),
		"sid":        sessionID,
		"iat":        now.Unix(),
		"exp":        now.Add(time.Hour * 300).Unix(),
	})

	accessStr, err := accessToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, err
	}
	refreshStr, err := refreshToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

