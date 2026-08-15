package auth

import (
	"errors"
	"strconv"

	"github.com/siddhantmadhur/pequod/config"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type authenticatedRoute func(echo.Context, *User, *config.Config) error

func getBearerTokenFromString(header string) (string, error) {
	rawBearerToken := header
	bearerToken := rawBearerToken[len("Bearer "):]
	if len(bearerToken) == 0 {
		return "", errors.New("Empty string provided")
	}
	return bearerToken, nil
}

// TODO: set permission level and profile picture
func AuthenticateRoute(next authenticatedRoute, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		rawBearerToken := c.Request().Header.Get("Authorization")
		bearerToken := rawBearerToken[len("Bearer "):]
		if bearerToken == "" {
			var result = map[string]string{
				"message": "Authoriation token not provided",
			}
			return c.JSON(500, result)
		}
		token, err := jwt.Parse(bearerToken, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.SecretKey), nil
		})
		if err != nil {
			var result = map[string]string{
				"message": "Authoriation token not provided",
				"error":   err.Error(),
			}
			return c.JSON(401, result)
		}
		if !token.Valid {
			var result = map[string]string{
				"message": "Token not valid",
			}
			return c.JSON(401, result)
		}

		claims := token.Claims.(jwt.MapClaims)

		var user User
		user.Username = claims["username"].(string)
		userId, err := strconv.Atoi(claims["uid"].(string))
		user.UID = int64(userId)
		return next(c, &user, cfg)
	}
}

func AuthenticateOrWizard(next authenticatedRoute, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !cfg.FinishedWizard {
			return next(c, nil, cfg)
		}
		return AuthenticateRoute(next, cfg)(c)
	}
}
