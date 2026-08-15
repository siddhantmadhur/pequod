package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/siddhantmadhur/pequod/internal/middleware"
)

type RouteConfig struct {
	Echo           *echo.Echo
	AuthHandler    *AuthHandler
	ServerHandler  *ServerHandler
	LibraryHandler *LibraryHandler
	MediaHandler   *MediaHandler
	AuthMiddleware *middleware.AuthMiddleware
}

func RegisterRoutes(cfg RouteConfig) {
	e := cfg.Echo
	authMw := cfg.AuthMiddleware

	// Wizard routes
	e.GET("/wizard/get-first-user", cfg.ServerHandler.GetWizardFirstUser, authMw.RequireWizardActive)

	// Server configuration
	e.GET("/server/information", cfg.ServerHandler.GetServerInformation)
	e.GET("/server/information/folders", cfg.ServerHandler.GetPathFolders, authMw.RequireAuthOrWizard)
	e.POST("/server/information/wizard", cfg.ServerHandler.FinishWizard, authMw.RequireWizardActive)

	// Library routes
	e.POST("/server/media/library", cfg.LibraryHandler.AddLibrary, authMw.RequireAuthOrWizard)
	e.GET("/server/media/library", cfg.LibraryHandler.GetLibraries, authMw.RequireAuthOrWizard)
	e.GET("/media/library/:mediaType/content", cfg.LibraryHandler.GetContentFromLibrary, authMw.RequireAuthOrWizard)
	e.GET("/media/library/:mediaId/children", cfg.LibraryHandler.GetVideoContentFromMedia, authMw.RequireAuthOrWizard)

	// Auth routes
	e.POST("/auth/create/user", cfg.AuthHandler.CreateUser, authMw.RequireAuthOrWizard)
	e.POST("/auth/login", cfg.AuthHandler.Login)
	e.POST("/auth/refresh", cfg.AuthHandler.RefreshToken)
	e.GET("/auth/user", cfg.AuthHandler.GetUser, authMw.RequireAuth)

	// Streaming & Playback routes
	e.POST("/media/:mediaId/playback/info", cfg.MediaHandler.GetPlaybackInfo, authMw.RequireAuth)
	e.GET("/media/:mediaId/streams/:sessionId/master.m3u8", cfg.MediaHandler.GetMasterPlaylist)
	e.GET("/media/:mediaId/streams/:sessionId/:segment/stream.ts", cfg.MediaHandler.GetStreamFile)
	e.GET("/media/:mediaId/direct/:fileName", cfg.MediaHandler.GetDirectPlayVideo, authMw.RequireAuth)
	e.GET("/server/streaming/sessions", cfg.MediaHandler.GetAllSessions, authMw.RequireAuth)
}
