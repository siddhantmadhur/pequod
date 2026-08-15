package auth

import (
	"github.com/siddhantmadhur/pequod/internal/config"

	"github.com/labstack/echo/v4"
)

type User struct {
	UID             int64  `json:"uid"`
	Username        string `json:"username"`
	AccessToken     string `json:"-"`
	RefreshToken    string `json:"-"`
	ProfilePicture  string `json:"profile_picture"`
	PermissionLevel int    `json:"-"`
}

func GetUser(c echo.Context, u *User, cfg *config.Config) error {
	return c.JSON(200, u)
}
