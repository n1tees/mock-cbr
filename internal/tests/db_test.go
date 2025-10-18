package tests

import (
	"database/sql"
	"testing"
	"time"

	"mock-cbr/internal/db"
)

func SetupTestDB(t *testing.T) *sql.DB {
	DB, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("cannot open test db: %v", err)
	}

	_, err = DB.Exec(`
		CREATE TABLE rates (
			date TEXT PRIMARY KEY,
			XML TEXT
		);
	`)
	if err != nil {
		t.Fatalf("cannot create table: %v", err)
	}

	start, _ := time.Parse("02/01/2006", "01/10/2025")
	end, _ := time.Parse("02/01/2006", "18/10/2025")

	if err := db.SeedRates(DB, start, end); err != nil {
		t.Fatalf("cannot seed test data: %v", err)
	}

	return DB
}
