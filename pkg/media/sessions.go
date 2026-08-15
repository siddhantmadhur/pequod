package media

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

func (m *Manager) GetAllSessions(c echo.Context) error {

	sessions := m.Sessions

	var result [](map[string]string)

	for k, v := range sessions {
		if v != nil {
			var temp = map[string]string{
				"session_id": k,
				"media_id":   fmt.Sprint(v.MediaId),
			}
			result = append(result, temp)
		}
	}

	return c.JSON(200, result)

}
