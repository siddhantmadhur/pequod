package main

import (
	"fmt"
	"log"
	"os"

	"github.com/siddhantmadhur/pequod/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	var cfg config.Config
	if os.Getenv("PERSISTENT_DATA") != "" {
		cfg.PersistentDir = os.Getenv("PERSISTENT_DATA")
	} else {
		cfg.PersistentDir = "/data"
	}
	err := cfg.Read()

	if err != nil {
		log.Fatal("[ERROR]: Config could not be read. " + err.Error())
		os.Exit(1)
	}

	err = runSqliteInit(&cfg)
	if err != nil {
		log.Fatal("There was an error in connecting to the sqlite file: " + err.Error())
		os.Exit(1)
	}

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		//AllowMethods: []string{"POST", "GET", "UPDATE", "DELETE"},
		AllowHeaders: []string{"Client-Name", "Content-Type", "Authorization"},
	}))

	handler(e, &cfg)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.Port)))
}
