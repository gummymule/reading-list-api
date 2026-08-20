package book

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrBookNotFound = errors.New("Book not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	repo := &Repository{
		db: db,
	}
	repo.seedIfEmpty()
	return repo
}

func (r *Repository) seedIfEmpty() {
	var count int
	r.db.QueryRow(`SELECT COUNT(*) FROM books`).Scan(&count)
	if count > 0 {
		return
	}
	cover1 := "https://m.media-amazon.com/images/I/51-1T3EnODL._SY445_SX342_FMwebp_.jpg"
	seedData := []Book{
		{
			ID:         uuid.NewString(),
			Title:      "Project Hail Mary",
			Author:     "Andy Weir",
			Genre:      "Science Fiction",
			CoverURL:   &cover1,
			Status:     StatusCurrentlyReading,
			Progress:   65,
			IsFavorite: false,
			AddedAt:    time.Now(),
		},
		{
			ID:         uuid.NewString(),
			Title:      "Educated",
			Author:     "Tara Westover",
			Genre:      "Memoir",
			Status:     StatusCurrentlyReading,
			Progress:   30,
			IsFavorite: true,
			AddedAt:    time.Now(),
		},
	}
	for _, b := range seedData {
		r.insert(b)
	}
}

func (r *Repository) insert(b Book) error {
	_, err := r.db.Exec(
		`INSERT INTO books (id, title, author, genre, cover_url, status, progress, is_favorite, added_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.Title, b.Author, b.Genre, b.CoverURL, b.Status, b.Progress, b.IsFavorite, b.AddedAt,
	)
	return err
}

var sortOptions = map[string]string{
	"newest":     "added_at DESC",
	"oldest":     "added_at ASC",
	"title-asc":  "title COLLATE NOCASE ASC",
	"title-desc": "title COLLATE NOCASE DESC",
}

func resolveSortClause(sort string) string {
	if clause, ok := sortOptions[sort]; ok {
		return clause
	}
	return sortOptions["newest"]
}

func (r *Repository) GetAll(sort string) ([]Book, error) {
	query := `SELECT id, title, author, genre, cover_url, status, progress, is_favorite, added_at
	FROM books ORDER BY ` + resolveSortClause(sort)
	return r.query(query)
}

func (r *Repository) GetByStatus(status ReadingStatus, sort string) ([]Book, error) {
	query := `SELECT id, title, author, genre, cover_url, status, progress, is_favorite, added_at
	FROM books WHERE status = ? ORDER BY ` + resolveSortClause(sort)
	return r.query(query, status)
}

func (r *Repository) GetFavorites(sort string) ([]Book, error) {
	query := `SELECT id, title, author, genre, cover_url, status, progress, is_favorite, added_at
	FROM books WHERE is_favorite = 1 ORDER BY ` + resolveSortClause(sort)
	return r.query(query)
}

func (r *Repository) query(query string, args ...any) ([]Book, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]Book, 0)
	for rows.Next() {
		var b Book
		var addedAt string
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Genre, &b.CoverURL, &b.Status, &b.Progress, &b.IsFavorite, &addedAt); err != nil {
			return nil, err
		}
		b.AddedAt, _ = time.Parse(time.RFC3339, addedAt)
		books = append(books, b)
	}
	return books, rows.Err()
}

func (r *Repository) Create(input CreateBookInput) (Book, error) {

	newBook := Book{
		ID:         uuid.NewString(),
		Title:      input.Title,
		Author:     input.Author,
		Genre:      input.Genre,
		CoverURL:   input.CoverURL,
		Status:     input.Status,
		Progress:   input.Progress,
		IsFavorite: false,
		AddedAt:    time.Now(),
	}
	err := r.insert(newBook)
	if err != nil {
		return Book{}, err
	}
	return newBook, nil
}

func (r *Repository) getByID(id string) (Book, error) {
	row := r.db.QueryRow(
		`SELECT id, title, author, genre, cover_url, status, progress, is_favorite, added_at
		FROM books WHERE id = ?`, id,
	)

	var b Book
	var addedAt string
	err := row.Scan(&b.ID, &b.Title, &b.Author, &b.Genre, &b.CoverURL, &b.Status, &b.Progress, &b.IsFavorite, &addedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return Book{}, ErrBookNotFound
	}
	if err != nil {
		return Book{}, err
	}
	b.AddedAt, _ = time.Parse(time.RFC3339, addedAt)
	return b, nil
}

func (r *Repository) Update(id string, input UpdateBookInput) (Book, error) {
	existing, err := r.getByID(id)
	if err != nil {
		return Book{}, err
	}

	if input.Title != nil {
		existing.Title = *input.Title
	}

	if input.Author != nil {
		existing.Author = *input.Author
	}

	if input.Genre != nil {
		existing.Genre = *input.Genre
	}

	if input.CoverURL != nil {
		existing.CoverURL = input.CoverURL
	}

	if input.Status != nil {
		existing.Status = *input.Status
	}

	if input.Progress != nil {
		existing.Progress = *input.Progress
	}

	_, err = r.db.Exec(
		`UPDATE books SET title = ?, author = ?, genre = ?, cover_url = ?, status = ?, progress = ? WHERE id = ?`,
		existing.Title, existing.Author, existing.Genre, existing.CoverURL, existing.Status, existing.Progress, id,
	)
	if err != nil {
		return Book{}, err
	}

	return existing, nil
}

func (r *Repository) ToggleFavorite(id string) (Book, error) {
	existing, err := r.getByID(id)
	if err != nil {
		return Book{}, err
	}

	newValue := !existing.IsFavorite
	_, err = r.db.Exec(`UPDATE books SET is_favorite = ? WHERE id = ?`, newValue, id)
	if err != nil {
		return Book{}, err
	}

	existing.IsFavorite = newValue
	return existing, nil
}

func (r *Repository) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM books WHERE id = ?`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrBookNotFound
	}

	return nil
}

func (r *Repository) CountByStatus(status ReadingStatus) int {
	var count int
	r.db.QueryRow(`SELECT COUNT(*) FROM books WHERE status = ?`, status).Scan(&count)
	return count
}
