package wizard

import (
	"net/http"

	"github.com/siddhantmadhur/pequod/internal/config"

	"github.com/labstack/echo/v4"
)

func WizardMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var config config.Config
		err := config.Read()
		if err != nil {
			return err
		}
		if config.FinishedWizard {
			return c.JSON(http.StatusUnauthorized, struct {
				Msg string `json:"msg"`
			}{
				Msg: "Server is already setup and you no longer have access to this feature",
			})
		}

		return next(c)
	}
}

func GetUser(c echo.Context) error {

	return c.String(200, "User")
}
