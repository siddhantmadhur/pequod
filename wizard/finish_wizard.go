package wizard

import (
	"github.com/siddhantmadhur/pequod/config"

	"github.com/labstack/echo/v4"
)

func FinishWizard(c echo.Context, config *config.Config) error {

	config.Mutex.Lock()
	defer config.Mutex.Unlock()

	config.FinishedWizard = true
	err := config.Write()
	if err != nil {
		return c.String(500, err.Error())
	}
	return c.NoContent(200)
}
