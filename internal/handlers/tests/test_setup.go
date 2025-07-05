package tests

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var testDB *sql.DB

// setupTestDB creates a new test database and returns it
func setupTestDB(t *testing.T) *sql.DB {
	// Use an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Error opening test database: %v", err)
	}

	// Create the items table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			value REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("Error creating test table: %v", err)
	}

	return db
}

func TestMain(m *testing.M) {
	// Suppress log output during tests
	log.SetOutput(os.NewFile(0, os.DevNull))

	// Run tests
	os.Exit(m.Run())
}
