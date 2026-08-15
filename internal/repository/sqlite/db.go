package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

// InitDB initializes SQLite database file, applies schema migrations, and configures connection pooling.
func InitDB(persistentDir string) (*sql.DB, error) {
	if err := os.MkdirAll(persistentDir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", persistentDir+"/storage.db")
	if err != nil {
		return nil, err
	}

	// SQLite performs best with controlled concurrency
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.ExecContext(context.Background(), ddl); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
