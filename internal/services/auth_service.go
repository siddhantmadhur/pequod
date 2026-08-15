package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, username, password, device, deviceName, clientName, clientVersion string) (*domain.TokenPair, error)
	RefreshToken(ctx context.Context, refreshTokenString string) (*domain.TokenPair, error)
	CreateUser(ctx context.Context, currentUser *domain.User, username, password string, permLevel int64, isWizardMode bool) error
	ValidateToken(tokenString string) (*domain.User, error)
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

	tokens, err := s.generateTokens(profile)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:               uuid.NewString(),
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
		// Log or handle session creation error if needed
	}

	return tokens, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshTokenString string) (*domain.TokenPair, error) {
	token, err := jwt.Parse(refreshTokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	profile, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.generateTokens(profile)
}

func (s *authService) CreateUser(ctx context.Context, currentUser *domain.User, username, password string, permLevel int64, isWizardMode bool) error {
	if username == "" || password == "" {
		return domain.ErrInvalidInput
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}

	if isWizardMode {
		count, err := s.userRepo.Count(ctx)
		if err != nil {
			return err
		}
		if count > 0 {
			admin, err := s.userRepo.GetAdmin(ctx)
			if err != nil {
				return err
			}
			return s.userRepo.Update(ctx, admin.ID, username, hashedPassword)
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
		return []byte(s.secretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	username, _ := claims["username"].(string)
	uidStr, _ := claims["uid"].(string)
	uid, _ := strconv.ParseInt(uidStr, 10, 64)

	return &domain.User{
		UID:          uid,
		Username:     username,
		AccessToken:  tokenString,
		RefreshToken: "",
	}, nil
}

func (s *authService) generateTokens(profile *domain.Profile) (*domain.TokenPair, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": profile.Username,
		"uid":      fmt.Sprint(profile.ID),
		"exp":      time.Now().Add(time.Minute * 20).Unix(),
	})
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": profile.Username,
		"uid":      fmt.Sprint(profile.ID),
		"exp":      time.Now().Add(time.Hour * 300).Unix(),
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
