package main

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed data.json
var seedData []byte

// Record represents a country entry in the database.
type Record struct {
	ID         int    `json:"id"`
	Country    string `json:"country"`
	Capital    string `json:"capital"`
	Population int    `json:"population"`
}

// Store defines the interface for data persistence.
// Implementations can use SQLite, PostgreSQL, Redis, etc.
type Store interface {
	GetAll() ([]Record, error)
	GetByID(id int) (*Record, error)
	Create(r Record) (*Record, error)
	Update(id int, r Record) (*Record, error)
	Delete(id int) error
	Count() (int, error)
	Close() error
}

// SQLiteStore implements Store using modernc.org/sqlite (pure Go, no CGO).
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) the SQLite database at dbPath.
// It creates the schema and seeds data if the table is empty.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	// Create schema
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS countries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			country TEXT NOT NULL,
			capital TEXT NOT NULL,
			population INTEGER NOT NULL
		)
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}

	s := &SQLiteStore{db: db}

	// Seed data if table is empty
	count, err := s.Count()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("count records: %w", err)
	}
	if count == 0 {
		if err := s.seed(); err != nil {
			db.Close()
			return nil, fmt.Errorf("seed data: %w", err)
		}
	}

	log.Printf("SQLiteStore ready: %s (%d records)", dbPath, count)
	return s, nil
}

func (s *SQLiteStore) seed() error {
	var records []Record
	if err := json.Unmarshal(seedData, &records); err != nil {
		return fmt.Errorf("unmarshal seed data: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO countries (country, capital, population) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, r := range records {
		if _, err := stmt.Exec(r.Country, r.Capital, r.Population); err != nil {
			return fmt.Errorf("insert %s: %w", r.Country, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Printf("Seeded %d records", len(records))
	return nil
}

func (s *SQLiteStore) GetAll() ([]Record, error) {
	rows, err := s.db.Query("SELECT id, country, capital, population FROM countries ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.ID, &r.Country, &r.Capital, &r.Population); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *SQLiteStore) GetByID(id int) (*Record, error) {
	var r Record
	err := s.db.QueryRow(
		"SELECT id, country, capital, population FROM countries WHERE id = ?", id,
	).Scan(&r.ID, &r.Country, &r.Capital, &r.Population)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *SQLiteStore) Create(r Record) (*Record, error) {
	result, err := s.db.Exec(
		"INSERT INTO countries (country, capital, population) VALUES (?, ?, ?)",
		r.Country, r.Capital, r.Population,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	r.ID = int(id)
	return &r, nil
}

func (s *SQLiteStore) Update(id int, r Record) (*Record, error) {
	result, err := s.db.Exec(
		"UPDATE countries SET country = ?, capital = ?, population = ? WHERE id = ?",
		r.Country, r.Capital, r.Population, id,
	)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, nil
	}
	r.ID = id
	return &r, nil
}

func (s *SQLiteStore) Delete(id int) error {
	result, err := s.db.Exec("DELETE FROM countries WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("record %d not found", id)
	}
	return nil
}

func (s *SQLiteStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM countries").Scan(&count)
	return count, err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
