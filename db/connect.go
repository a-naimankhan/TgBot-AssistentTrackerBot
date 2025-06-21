package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func runMigrations(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY,
			username TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS words (
			id SERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
			word TEXT NOT NULL,
			translation TEXT NOT NULL,
			deadline TIMESTAMP,
			added_at TIMESTAMP DEFAULT NOW(),
			correct_count INT DEFAULT 0,
			is_learned BOOLEAN DEFAULT FALSE
		);`,
		`CREATE TABLE IF NOT EXISTS goals (
			goal_id SERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
			goal VARCHAR(255) NOT NULL,
			deadline TIMESTAMP,
			added_at TIMESTAMP DEFAULT NOW(),
			completed BOOLEAN DEFAULT FALSE
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func InitDB() error {
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		log.Fatal("❌ DB_URL is not set")
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("❌ Failed to open DB: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("❌ Failed to connect to DB: %w", err)
	}

	log.Println("✅ DB is connected")

	// Теперь можно запускать миграции
	if err := runMigrations(DB); err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	return nil
}
