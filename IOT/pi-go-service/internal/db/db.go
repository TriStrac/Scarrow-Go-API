package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	HubID string
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, error := sql.Open("sqlite3", dbPath)
	if error != nil {
		return nil, error
	}

	// Create table if not exists
	query := `
	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT
	);`
	_, error = db.Exec(query)
	if error != nil {
		return nil, error
	}

	return db, nil
}

func GetHubID(db *sql.DB) (string, error) {
	var hubID string
	err := db.QueryRow("SELECT value FROM config WHERE key = 'hub_id'").Scan(&hubID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return hubID, nil
}

func GetHubSecret(db *sql.DB) (string, error) {
	var secret string
	err := db.QueryRow("SELECT value FROM config WHERE key = 'hub_secret'").Scan(&secret)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return secret, nil
}

func SaveConfig(db *sql.DB, key, value string) error {
	_, err := db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", key, value)
	return err
}
