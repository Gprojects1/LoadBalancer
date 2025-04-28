package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func Connect(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS clients (
            client_id TEXT PRIMARY KEY,
            capacity INTEGER NOT NULL,
            rate_per_sec DOUBLE PRECISION NOT NULL
        );
    `)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}
