package user

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("User not found")
var ErrEmailTaken = errors.New("Email already registered")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(email, passwordHash, name string) (User, error) {
	var exists int
	r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, email).Scan(&exists)
	if exists > 0 {
		return User{}, ErrEmailTaken
	}

	newUser := User{
		ID:        uuid.NewString(),
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}

	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password_hash, name, created_at) VALUES (?, ?, ?, ?, ?)`,
		newUser.ID, newUser.Email, passwordHash, newUser.Name, newUser.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return User{}, err
	}
	return newUser, nil
}

func (r *Repository) GetByEmailWithHash(email string) (User, string, error) {
	row := r.db.QueryRow(
		`SELECT id, email, password_hash, name, created_at FROM users WHERE email = ?`, email,
	)
	var user User
	var passwordHash, createdAt string
	err := row.Scan(&user.ID, &user.Email, &passwordHash, &user.Name, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, "", ErrUserNotFound
	}
	if err != nil {
		return User{}, "", err
	}

	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return user, passwordHash, nil
}

func (r *Repository) GetByID(id string) (User, error) {
	row := r.db.QueryRow(`SELECT id, email, name, created_at FROM users WHERE id = ?`, id)

	var user User
	var createdAt string
	err := row.Scan(&user.ID, &user.Email, &user.Name, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}

	if err != nil {
		return User{}, err
	}

	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return user, nil
}
