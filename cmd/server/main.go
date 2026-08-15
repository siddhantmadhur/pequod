package main

import (
	"log"
	"os"

	"github.com/siddhantmadhur/pequod/internal/config"
	"github.com/siddhantmadhur/pequod/internal/repository/sqlite"
	"github.com/siddhantmadhur/pequod/internal/server"
	"github.com/siddhantmadhur/pequod/internal/services"
)

func main() {
	var cfg config.Config
	if os.Getenv("PERSISTENT_DATA") != "" {
		cfg.PersistentDir = os.Getenv("PERSISTENT_DATA")
	} else {
		cfg.PersistentDir = "/data"
	}

	if err := cfg.Read(); err != nil {
		log.Fatalf("[ERROR]: Config could not be read: %v", err)
	}

	// 1. Initialize SQLite Database Connection Pool
	db, err := sqlite.InitDB(cfg.PersistentDir)
	if err != nil {
		log.Fatalf("[ERROR]: Failed to connect to SQLite database: %v", err)
	}
	defer db.Close()

	// 2. Initialize Repository Layer
	userRepo := sqlite.NewUserRepository(db)
	sessionRepo := sqlite.NewSessionRepository(db)
	libraryRepo := sqlite.NewLibraryRepository(db)
	contentRepo := sqlite.NewContentRepository(db)

	// 3. Initialize Service Layer
	authService := services.NewAuthService(userRepo, sessionRepo, cfg.SecretKey)
	serverService := services.NewServerService(&cfg)
	scannerService := services.NewScannerService(contentRepo)
	libraryService := services.NewLibraryService(libraryRepo, contentRepo, userRepo, scannerService)
	mediaService := services.NewMediaService(contentRepo, &cfg)

	// 4. Initialize HTTP Presentation Layer & Server
	srv := server.New(server.ServerParams{
		Config:         &cfg,
		AuthService:    authService,
		ServerService:  serverService,
		LibraryService: libraryService,
		MediaService:   mediaService,
	})

	log.Printf("[INFO]: Starting Pequod Media Server on port %d...", cfg.Port)
	srv.Echo.Logger.Fatal(srv.Start())
}
