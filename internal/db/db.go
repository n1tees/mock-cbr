package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
