package goal

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetTarget(userID string, year int) int {
	var target int
	err := r.db.QueryRow(`SELECT target FROM reading_goals WHERE user_id = ? AND year = ?`, userID, year).Scan(&target)

	if err != nil {
		return 0
	}

	return target
}

func (r *Repository) SetTarget(userID string, year int, target int) error {
	_, err := r.db.Exec(
		`INSERT INTO reading_goals (user_id, year, target) VALUES (?, ?, ?)
		ON CONFLICT(user_id, year) DO UPDATE SET target = excluded.target`,
		userID, year, target,
	)
	return err
}
