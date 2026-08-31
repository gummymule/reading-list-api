package user

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("User not found")
var ErrEmailTaken = errors.New("Email already registered")
var ErrInvalidOTP = errors.New("Invalid or expired code")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func generateOTP() (string, error) {
	max := big.NewInt(1000000) // 6 digits
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (r *Repository) CreatePasswordReset(email string) (string, error) {
	var exists int
	r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, email).Scan(&exists)
	if exists == 0 {
		return "", ErrUserNotFound
	}

	otp, err := generateOTP()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(10 * time.Minute)

	_, err = r.db.Exec(
		`INSERT OR REPLACE INTO password_resets (email, otp_code, expires_at) VALUES (?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET otp_code = excluded.otp_code, expires_at = excluded.expires_at`,
		email, otp, expiresAt.Format(time.RFC3339),
	)
	if err != nil {
		return "", err
	}

	return otp, nil
}

func (r *Repository) ResetPassword(email, otp, newPasswordHash string) error {
	var storedOTP, expiresAtStr string
	err := r.db.QueryRow(
		`SELECT otp_code, expires_at FROM password_resets WHERE email = ?`, email,
	).Scan(&storedOTP, &expiresAtStr)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidOTP
	}
	if err != nil {
		return err
	}

	expiresAt, _ := time.Parse(time.RFC3339, expiresAtStr)
	if time.Now().After(expiresAt) {
		return ErrInvalidOTP
	}
	if storedOTP != otp {
		return ErrInvalidOTP
	}

	_, err = r.db.Exec(`UPDATE users SET password_hash = ? WHERE email = ?`, newPasswordHash, email)
	if err != nil {
		return err
	}

	r.db.Exec(`DELETE FROM password_resets WHERE email = ?`, email)

	return nil
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
