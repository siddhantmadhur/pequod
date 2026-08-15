package server

import (
	"fmt"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/siddhantmadhur/pequod/internal/config"
	"github.com/siddhantmadhur/pequod/internal/handlers"
	"github.com/siddhantmadhur/pequod/internal/middleware"
)

type Server struct {
	Echo   *echo.Echo
	Config *config.Config
}

func New(cfg *config.Config) *Server {
	e := echo.New()

	e.Use(middleware.Logger)

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		//AllowMethods: []string{"POST", "GET", "UPDATE", "DELETE"},
		AllowHeaders: []string{"Client-Name", "Content-Type", "Authorization"},
	}))

	handlers.RegisterRoutes(e, cfg)

	return &Server{
		Echo:   e,
		Config: cfg,
	}
}

func (s *Server) Start() error {
	return s.Echo.Start(fmt.Sprintf(":%d", s.Config.Port))
}
