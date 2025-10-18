package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func InitDB(path string) (*sql.DB, error) {

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory for DB: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("DB not found, creating new - %s", path)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database - %s", path)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database - %v", err)
	}

	createTable := `
    CREATE TABLE IF NOT EXISTS rates (
        date TEXT PRIMARY KEY,
        xml TEXT NOT NULL
    );`

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, fmt.Errorf("failed to create table - %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM rates").Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check records %v", err)
	}

	if count == 0 {
		log.Println("DB is empty, seeding with test data...")

		start, _ := time.Parse("02/01/2006", "01/01/2024")
		end, _ := time.Parse("02/01/2006", "31/12/2025")

		if err := SeedRates(db, start, end); err != nil {
			return nil, fmt.Errorf("failed to seed data - %v", err)
		}
	} else {
		log.Printf("Connect to databse is done, has %d records", count)
	}

	return db, nil
}

func GetRate(db *sql.DB, date string) (string, error) {

	var xml string

	err := db.QueryRow("SELECT XML FROM rates WHERE DATE = ?", date).Scan(&xml)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no data for date %s", date)
	}

	return xml, err
}
