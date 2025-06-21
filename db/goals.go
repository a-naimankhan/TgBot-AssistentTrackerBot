package db

import (
	"fmt"
	"time"
)

func AddGoal(chatId int64, goal string, deadline time.Time, remainder int) error {
	_, err := DB.Exec(`
		INSERT INTO goals (user_id, goal, deadline, completed)
		VALUES ($1, $2, $3,  , $4)
	`, chatId, goal, deadline, nil)
	return err
}

func GetGoals(userID int64) ([]string, error) {
	rows, err := DB.Query(`SELECT goal, deadline FROM goals WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []string
	for rows.Next() {
		var goal string
		var deadline time.Time
		if err := rows.Scan(&goal, &deadline); err != nil {
			continue
		}
		goals = append(goals, fmt.Sprintf("🎯 %s — до %s", goal, deadline.Format("02 Jan 2006")))
	}

	return goals, nil
}
