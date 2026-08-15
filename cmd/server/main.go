package main

import (
	"log"
	"os"

	"github.com/siddhantmadhur/pequod/internal/config"
	"github.com/siddhantmadhur/pequod/internal/database"
	"github.com/siddhantmadhur/pequod/internal/server"
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

	err = database.RunSqliteInit(&cfg)
	if err != nil {
		log.Fatal("There was an error in connecting to the sqlite file: " + err.Error())
		os.Exit(1)
	}

	srv := server.New(&cfg)
	srv.Echo.Logger.Fatal(srv.Start())
}
