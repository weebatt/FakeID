package database

import (
	"database/sql"
	_ "github.com/lib/pq"
	"time"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(connectionString string) (*Database, error) {
	// Open database connection
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)                 // Limit maximum simultaneous connections
	db.SetMaxIdleConns(5)                  // Keep some connections ready
	db.SetConnMaxLifetime(5 * time.Minute) // Refresh connections periodically

	// Verify connection is working
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}
