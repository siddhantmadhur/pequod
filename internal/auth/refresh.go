package auth

import (
	"context"
	"net/http"

	"github.com/siddhantmadhur/pequod/internal/config"
	"github.com/siddhantmadhur/pequod/internal/database"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func RefreshToken(c echo.Context, cfg *config.Config) error {
	bearerToken, err := getBearerTokenFromString(c.Request().Header.Get("Authorization"))
	if err != nil {
		var result = map[string]string{
			"message": "Could not store token in database",
			"error":   err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, result)
	}
	token, err := jwt.Parse(bearerToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.SecretKey), nil
	})
	if err != nil {
		var result = map[string]string{
			"message": "Authoriation token not provided",
			"error":   err.Error(),
		}
		return c.JSON(500, result)
	}
	if !token.Valid {
		var result = map[string]string{
			"message": "Token not valid",
		}
		return c.JSON(401, result)
	}

	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	conn, query, err := database.GetConn(cfg)
	if err != nil {
		var result = map[string]string{
			"message": "There was an error with connecting to the database",
			"error":   err.Error(),
		}
		return c.JSON(500, result)
	}
	defer conn.Close()

	user, err := query.GetUserFromUsername(context.Background(), username)
	if err != nil {
		var result = map[string]string{
			"message": "There was an error with fetching the user",
			"error":   err.Error(),
		}
		return c.JSON(500, result)
	}

	accessToken, refreshToken, err := generateToken(&user, cfg.SecretKey)
	var result = map[string]interface{}{
		"message": "Refresh successful",
		"data": map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	}

	return c.JSON(200, result)
}
