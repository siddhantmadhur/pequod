package middleware

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
)

// Use this middleware to log HTTP requests
func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cur := time.Now()
		res := next(c)
		don := time.Since(cur)

		log.Printf("[%s] %d - %s (%s) \n", c.Request().Method, c.Response().Status, c.Request().RequestURI, don.String())

		return res
	}
}
