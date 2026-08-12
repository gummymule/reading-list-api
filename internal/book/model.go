package book

import "time"

type ReadingStatus string

const (
	StatusWantToRead       ReadingStatus = "want-to-read"
	StatusCurrentlyReading ReadingStatus = "currently-reading"
	StatusRead             ReadingStatus = "read"
)

type Book struct {
	ID         string        `json:"id"`
	Title      string        `json:"title"`
	Author     string        `json:"author"`
	Genre      string        `json:"genre"`
	CoverURL   *string       `json:"coverUrl,omitempty"`
	Status     ReadingStatus `json:"status"`
	Progress   int           `json:"progress"`
	IsFavorite bool          `json:"isFavorite"`
	AddedAt    time.Time     `json:"AddedAt"`
}

type CreateBookInput struct {
	Title    string        `json:"title"`
	Author   string        `json:"author"`
	Genre    string        `json:"genre"`
	CoverURL *string       `json:"coverUrl"`
	Status   ReadingStatus `json:"status"`
	Progress int           `json:"progress"`
}

type UpdateBookInput struct {
	Title    *string        `json:"title,omitempty"`
	Author   *string        `json:"author,omitempty"`
	Genre    *string        `json:"genre,omitempty"`
	CoverURL *string        `json:"coverUrl,omitempty"`
	Status   *ReadingStatus `json:"status,omitempty"`
	Progress *int           `json:"progress,omitempty"`
}
