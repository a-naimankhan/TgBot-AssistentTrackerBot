package db

import (
	"fmt"
	"time"
)

func AddWords(id int64, word, translation string, deadline time.Time) error {
	_, err := DB.Exec(`
	INSERT INTO words (user_id , word , translation , deadline , added_at)
	VALUES ($1 , $2 , $3 , $4 , $5)
	
	`, id, word, translation, deadline, time.Now())
	return err
}

func GetUserWords(userid int64) ([]string, error) {
	rows, err := DB.Query(`SELECT word, translation FROM words WHERE user_id = $1`, userid)
	if err != nil {
		return nil, err
	}

	var words []string
	for rows.Next() {
		var word, translation string
		if err := rows.Scan(&word, &translation); err != nil {
			continue
		}
		words = append(words, fmt.Sprintf("%s ------ %s", word, translation))
	}
	return words, nil
}

func GetRandomWord(userid int64) (string, string, int, error) {
	row := DB.QueryRow(
		`SELECT word, translation , correct_count
		FROM words
		WHERE user_id = ($1)
		ORDER BY RANDOM() 
		LIMIT 1`, userid)
	var word, translation string
	var count int
	err := row.Scan(&word, &translation, &count)
	if err != nil {
		return "", "", 0, err
	}

	return word, translation, count, nil
}
