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
		WHERE user_id = ($1) and is_learned = false
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

func UpdateWordCorrectCount(userID int64, word string, count int, isLearned bool) error {
	_, err := DB.Exec(`
		UPDATE words 
		SET correct_count = $1, is_learned = $2
		WHERE user_id = $3 AND word = $4`,
		count, isLearned, userID, word)
	return err
}

func SetUserState(userID int64, word, correctAns string, count int) error {
	_, err := DB.Exec(`
		INSERT INTO user_state (user_id, current_word, correct_answer, correct_count)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET current_word = $2, correct_answer = $3, correct_count = $4
	`, userID, word, correctAns, count)
	return err
}

type UserState struct {
	Word       string
	CorrectAns string
	Count      int
}

func GetUserState(userID int64) (*UserState, error) {
	row := DB.QueryRow(`
		SELECT current_word, correct_answer, correct_count
		FROM user_state
		WHERE user_id = $1
	`, userID)

	var state UserState
	err := row.Scan(&state.Word, &state.CorrectAns, &state.Count)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

/* func AddNewTable() {
	_, err := DB.Exec(`CREATE TABLE user_state (
    user_id BIGINT PRIMARY KEY,
    current_word TEXT,
    correct_answer TEXT,
    correct_count INT DEFAULT 0
);`)
	if err != nil {
		log.Print("Error")
	}
	log.Println("New table added")

}
*/
