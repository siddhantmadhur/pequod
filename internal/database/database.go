package database

import (
	"context"
	"database/sql"
	_ "embed"
	"os"

	"github.com/siddhantmadhur/pequod/internal/config"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

func RunSqliteInit(cfg *config.Config) error {
	ctx := context.Background()
	os.MkdirAll(cfg.PersistentDir, 0755)
	db, err := sql.Open("sqlite3", cfg.PersistentDir+"/storage.db")
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.ExecContext(ctx, ddl)
	if err != nil {
		return err
	}
	return nil
}
