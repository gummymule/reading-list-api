package book

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrBookNotFound = errors.New("Book not found")

type Repository struct {
	mu    sync.RWMutex
	books map[string]Book
}

func NewRepository() *Repository {
	repo := &Repository{
		books: make(map[string]Book),
	}
	repo.seed()
	return repo
}

func (r *Repository) seed() {
	cover1 := "https://covers.openlibrary.org/b/id/12003713-L.jpg"
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
		r.books[b.ID] = b
	}
}

func (r *Repository) GetAll() []Book {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Book, 0, len(r.books))
	for _, b := range r.books {
		result = append(result, b)
	}
	return result
}

func (r *Repository) GetByStatus(status ReadingStatus) []Book {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Book, 0)
	for _, b := range r.books {
		if b.Status == status {
			result = append(result, b)
		}
	}
	return result
}

func (r *Repository) GetFavorites() []Book {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Book, 0)
	for _, b := range r.books {
		if b.IsFavorite {
			result = append(result, b)
		}
	}
	return result
}

func (r *Repository) Create(input CreateBookInput) Book {
	r.mu.Lock()
	defer r.mu.Unlock()

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
	r.books[newBook.ID] = newBook
	return newBook
}

func (r *Repository) Update(id string, input UpdateBookInput) (Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.books[id]
	if !ok {
		return Book{}, ErrBookNotFound
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

	if input.Status != nil {
		existing.Status = *input.Status
	}

	if input.Progress != nil {
		existing.Progress = *input.Progress
	}

	r.books[id] = existing
	return existing, nil
}

func (r *Repository) ToggleFavorite(id string) (Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.books[id]
	if !ok {
		return Book{}, ErrBookNotFound
	}

	existing.IsFavorite = !existing.IsFavorite
	r.books[id] = existing
	return existing, nil
}

func (r *Repository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.books[id]; !ok {
		return ErrBookNotFound
	}
	delete(r.books, id)
	return nil
}
