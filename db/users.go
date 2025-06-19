package db

import (
	"fmt"
	"time"
)

func AddUser(id int64, username string) error {
	fmt.Printf("👤 Добавление пользователя: ID=%d, Username=%s\n", id, username)

	_, err := DB.Exec(`
		INSERT INTO users (id, username, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, id, username, time.Now())
	return err
}
