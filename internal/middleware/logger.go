package middleware

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
)

// Logger provides structured request logging for Echo HTTP routes.
func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		latency := time.Since(start)

		log.Printf("[%s] %d - %s (%s) \n", c.Request().Method, c.Response().Status, c.Request().RequestURI, latency.String())
		return err
	}
}
