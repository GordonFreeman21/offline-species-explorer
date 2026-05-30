package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Species struct {
	ID             int    `json:"id"`
	CommonName     string `json:"common_name"`
	ScientificName string `json:"scientific_name"`
	Kingdom        string `json:"kingdom"`
	Phylum         string `json:"phylum"`
	Class          string `json:"class"`
	Order          string `json:"order"`
	Family         string `json:"family"`
	LastSynced     string `json:"last_synced"`
}

var db *sql.DB

func InitDB() error {
	dbPath, err := getDBPath()
	if err != nil {
		return fmt.Errorf("getting db path: %w", err)
	}

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := createTable(); err != nil {
		return fmt.Errorf("creating table: %w", err)
	}

	count, err := GetSpeciesCount()
	if err != nil {
		return fmt.Errorf("counting species: %w", err)
	}

	if count == 0 {
		if err := SeedDatabase(db); err != nil {
			return fmt.Errorf("seeding database: %w", err)
		}
	}

	return nil
}

func getDBPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}

	appDir := filepath.Join(configDir, "offline-species-explorer")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("creating app directory: %w", err)
	}

	return filepath.Join(appDir, "species.db") + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", nil
}

func createTable() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS species (
			id INTEGER PRIMARY KEY,
			common_name TEXT,
			scientific_name TEXT UNIQUE NOT NULL,
			kingdom TEXT,
			phylum TEXT,
			class TEXT,
			"order" TEXT,
			family TEXT,
			last_synced DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func GetSpeciesCount() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM species").Scan(&count)
	return count, err
}

func UpdateSpecies(s Species) error {
	stmt, err := db.Prepare(`
		UPDATE species SET
			common_name = ?,
			kingdom = ?,
			phylum = ?,
			class = ?,
			"order" = ?,
			family = ?,
			last_synced = ?
		WHERE scientific_name = ?
	`)
	if err != nil {
		return fmt.Errorf("preparing update: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = stmt.Exec(s.CommonName, s.Kingdom, s.Phylum, s.Class, s.Order, s.Family, now, s.ScientificName)
	return err
}

func InsertSpecies(s Species) (int64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO species (common_name, scientific_name, kingdom, phylum, class, "order", family, last_synced)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("preparing insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := stmt.Exec(s.CommonName, s.ScientificName, s.Kingdom, s.Phylum, s.Class, s.Order, s.Family, now)
	if err != nil {
		return 0, fmt.Errorf("inserting species: %w", err)
	}

	return result.LastInsertId()
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}
