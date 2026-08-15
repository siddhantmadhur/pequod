package main

import (
	"log"

	"github.com/siddhantmadhur/pequod/auth"
	"github.com/siddhantmadhur/pequod/config"
	"github.com/siddhantmadhur/pequod/library"
	"github.com/siddhantmadhur/pequod/media"
	"github.com/siddhantmadhur/pequod/wizard"

	"github.com/labstack/echo/v4"
)

func handler(e *echo.Echo, cfg *config.Config) {

	e.Use(Logger)

	//var ffmpeg = media.NewFfmpeg("veryfast")

	// Wizard routes
	e.GET("/wizard/get-first-user", wizard.WizardMiddleware(wizard.GetUser))

	// Server config
	e.GET("/server/information", cfg.GetServerInformation)
	e.GET("/server/information/folders", auth.AuthenticateOrWizard(library.GetPathFolders, cfg))
	e.POST("/server/information/wizard", cfg.Route(wizard.FinishWizard))

	// Library
	e.POST("/server/media/library", auth.AuthenticateOrWizard(library.AddLibraryFolder, cfg))
	e.GET("/server/media/library", auth.AuthenticateOrWizard(library.GetLibraryFolders, cfg))

	e.GET("/media/library/:mediaType/content", cfg.Route(library.GetContentFromLibrary))
	e.GET("/media/library/:mediaId/children", auth.AuthenticateOrWizard(library.GetVideoContentFromMedia, cfg))

	// Auth routes
	e.POST("/auth/create/user", auth.AuthenticateOrWizard(auth.CreateNewUserRoute, cfg))
	e.POST("/auth/login", cfg.Route(auth.Login))
	e.POST("/auth/refresh", cfg.Route(auth.RefreshToken))
	e.GET("/auth/user", auth.AuthenticateRoute(auth.GetUser, cfg))

	// Streaming routes
	streamer, err := media.NewManager(cfg)
	if err != nil {
		log.Fatalf("[ERROR]: %s\n", err.Error())
	}

	// Creates the right m3u url for the playback client. i.e. what time to resume, subtitles to use etc.
	e.POST("/media/:mediaId/playback/info", auth.AuthenticateRoute(streamer.GetPlaybackInfo, cfg))

	// Once the m3u8 url is made it will be in the format below
	e.GET("/media/:mediaId/streams/:sessionId/master.m3u8", streamer.GetMasterPlaylist)

	// The m3u8 will refer to segments from below
	// /media/:mediaId/streams/:sessionId/:segment/stream.ts
	e.GET("/media/:mediaId/streams/:sessionId/:segment/stream.ts", streamer.GetStreamFile)

	e.GET("/media/:mediaId/direct/:fileName", cfg.Route(streamer.GetDirectPlayVideo))

	e.GET("/server/streaming/sessions", streamer.GetAllSessions)
}
