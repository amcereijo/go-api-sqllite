package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite database connection
func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	err = createTables(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// createTables creates the necessary database tables
func createTables(db *sql.DB) error {
	// Example table creation - modify according to your needs
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS features (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		value TEXT NOT NULL,
		resource_id TEXT NOT NULL,
		active BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createTableSQL)
	return err
}
