package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/siddhantmadhur/pequod/config"
	"github.com/siddhantmadhur/pequod/storage"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func generateToken(user *storage.Profile, secret string) (string, string, error) {

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"uid":      fmt.Sprint(user.ID),
		"exp":      time.Now().Add(time.Minute * 20).Unix(),
	})
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"uid":      fmt.Sprint(user.ID),
		"exp":      time.Now().Add(time.Hour * 300).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(secret))
	refreshTokenString, err := refreshToken.SignedString([]byte(secret))

	return accessTokenString, refreshTokenString, err
}

func storeToken(userId int64, accessToken string, refreshToken string, device string, deviceName string, clientName string, clientVersion string, query *storage.Queries) (storage.Session, error) {
	return query.CreateSession(context.Background(), storage.CreateSessionParams{
		ID:               uuid.NewString(),
		UserID:           userId,
		CreatedAt:        time.Now(),
		AccessExpiresAt:  time.Now().Add(time.Minute * 20),
		RefreshExpiresAt: time.Now().Add(time.Hour * 300),
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		Device:           device,
		DeviceName:       deviceName,
		ClientName:       clientName,
		ClientVersion:    clientVersion,
	})
}

func Login(c echo.Context, cfg *config.Config) error {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := c.Bind(&request)
	if err != nil {
		var result = map[string]string{
			"message": "There was an error",
			"error":   err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	conn, query, err := storage.GetConn(cfg)
	defer conn.Close()
	if err != nil {
		var result = map[string]string{
			"message": "There was an error",
			"error":   err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	user, err := comparePasswords(query, request.Username, request.Password)
	if err != nil {
		var result = map[string]string{
			"message": "Password does not match or there was an error",
			"error":   err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	// TODO: Create session/jwt whatever and send to the user
	accessToken, refreshToken, err := generateToken(user, cfg.SecretKey)
	if err != nil {
		var result = map[string]string{
			"message": "Could not generate JWT",
			"error":   err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	// TODO: Change to device and client information
	//tokenInformation, err := storeToken(user.ID, accessToken, refreshToken, "", "", "", "", query)
	//if err != nil {
	//var result = map[string]string{
	//"message": "Could not store token in database",
	//"error":   err.Error(),
	//	}
	//	return c.JSON(http.StatusInternalServerError, result)
	//}

	var result = map[string]interface{}{
		"message": "Login successful",
		"data": map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	}

	return c.JSON(200, result)
}

func comparePasswords(query *storage.Queries, username string, password string) (*storage.Profile, error) {

	user, err := query.GetUserFromUsername(context.Background(), username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	return &user, err
}
