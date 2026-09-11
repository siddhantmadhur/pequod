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

	dsn := persistentDir + "/storage.db?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on"
	db, err := sql.Open("sqlite3", dsn)
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
