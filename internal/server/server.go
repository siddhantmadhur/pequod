package server

import (
	"fmt"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/siddhantmadhur/pequod/internal/config"
	"github.com/siddhantmadhur/pequod/internal/handlers"
	"github.com/siddhantmadhur/pequod/internal/middleware"
	"github.com/siddhantmadhur/pequod/internal/services"
)

type Server struct {
	Echo   *echo.Echo
	Config *config.Config
}

type ServerParams struct {
	Config         *config.Config
	AuthService    services.AuthService
	ServerService  services.ServerService
	LibraryService services.LibraryService
	MediaService   services.MediaService
}

func New(params ServerParams) *Server {
	e := echo.New()

	e.Use(middleware.Logger)
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowHeaders: []string{"Client-Name", "Content-Type", "Authorization"},
	}))

	authMw := middleware.NewAuthMiddleware(params.AuthService, params.ServerService)

	authHandler := handlers.NewAuthHandler(params.AuthService, params.ServerService)
	serverHandler := handlers.NewServerHandler(params.ServerService, params.LibraryService)
	libraryHandler := handlers.NewLibraryHandler(params.LibraryService)
	mediaHandler := handlers.NewMediaHandler(params.MediaService)

	handlers.RegisterRoutes(handlers.RouteConfig{
		Echo:           e,
		AuthHandler:    authHandler,
		ServerHandler:  serverHandler,
		LibraryHandler: libraryHandler,
		MediaHandler:   mediaHandler,
		AuthMiddleware: authMw,
	})

	return &Server{
		Echo:   e,
		Config: params.Config,
	}
}

func (s *Server) Start() error {
	return s.Echo.Start(fmt.Sprintf(":%d", s.Config.Port))
}
