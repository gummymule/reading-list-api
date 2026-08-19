package goal

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetTarget(year int) int {
	var target int
	err := r.db.QueryRow(`SELECT target FROM reading_goals WHERE year = ?`, year).Scan(&target)

	if err != nil {
		return 0
	}

	return target
}

func (r *Repository) SetTarget(year int, target int) error {
	_, err := r.db.Exec(
		`INSERT INTO reading_goals (year, target) VALUES (?, ?)
		ON CONFLICT(year) DO UPDATE SET target = excluded.target`,
		year, target,
	)
	return err
}
