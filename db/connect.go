package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	connStr := os.Getenv("DB_URL")
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	fmt.Println("DB is connected :3")
	return DB.Ping()
}
