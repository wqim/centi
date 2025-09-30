package util

import (
	"database/sql"
	"net/url"
	"fmt"
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
	_ "github.com/xeodou/go-sqlcipher"
)

func TestConnectDB(t *testing.T) {
	// Create a temporary file for testing.  Crucial for avoiding conflicts.
	tempFile, err := os.CreateTemp("", "test_db_*.db")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up after the test

	password := "mysecretpassword" // Replace with a strong password for production

	db, err := ConnectDB(tempFile.Name(), password, 100)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close() // Important to close the connection

	if err = db.InitDB(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Assert the database connection is valid
	assert.NoError(t, err, "Database connection should be successful")
	assert.NotNil(t, db, "Database object should not be nil")
	assert.Equal(t, db.rowsLimit, uint(100), "Rows limit should be correctly set")

	// Test counting rows (initial count should be 0)
	rows, err := db.Count()
	assert.NoError(t, err, "Count operation should succeed")
	assert.Equal(t, rows, 0, "Initial row count should be 0")
}

// Helper function (you might want to put this in a separate file if it's reused)
func createTestDB() (*sql.DB, string, error) {
	tempFile, err := os.CreateTemp("", "test_db_*.db")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp file: %w", err)
	}
	password := "mysecretpassword"
	dbFilename := "file:" + url.QueryEscape(tempFile.Name()) + "?_journal_mode=WAL&_key=" + url.QueryEscape(password)
	db, err := sql.Open("sqlite3", dbFilename)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open DB: %w", err)
	}
	return db, tempFile.Name(), nil
}

func TestShredFile(t *testing.T) {
	// Create a temporary file for testing.
	tempFile, err := os.CreateTemp("", "test_db_*.db")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Simulate shredding the file.  This is a critical test!  You need to
	// actually call ShredFile with the correct implementation.
	err = ShredFile(tempFile.Name())
	assert.NoError(t, err, "File shredding should succeed")

	// Verify the file is gone (or significantly altered).  Crucial!
	_, err = os.Stat(tempFile.Name())
	assert.Error(t, err, "File should no longer exist")
}
