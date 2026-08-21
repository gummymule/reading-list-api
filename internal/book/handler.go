package book

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"reading-list-api/internal/auth"
	"reading-list-api/internal/response"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/books", auth.RequireAuth(http.HandlerFunc(h.list)))
	mux.Handle("POST /api/books", auth.RequireAuth(http.HandlerFunc(h.create)))
	mux.Handle("PATCH /api/books/{id}", auth.RequireAuth(http.HandlerFunc(h.update)))
	mux.Handle("PATCH /api/books/{id}/favorite", auth.RequireAuth(http.HandlerFunc(h.toggleFavorite)))
	mux.Handle("DELETE /api/books/{id}", auth.RequireAuth(http.HandlerFunc(h.delete)))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	status := r.URL.Query().Get("status")
	favorite := r.URL.Query().Get("favorite")
	search := r.URL.Query().Get("search")
	sort := r.URL.Query().Get("sort")

	var books []Book
	var err error

	switch {
	case favorite == "true":
		books, err = h.repo.GetFavorites(userID, sort)
	case status != "":
		books, err = h.repo.GetByStatus(userID, ReadingStatus(status), sort)
	default:
		books, err = h.repo.GetAll(userID, sort)
	}

	if err != nil {
		log.Printf("list books error: %v", err)
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Failed to fetch books")
		return
	}

	if search != "" {
		books = filterBySearch(books, search)
	}

	response.Success(w, http.StatusOK, "Success Get Books", books)
}

func filterBySearch(books []Book, query string) []Book {
	search := strings.ToLower(query)
	result := make([]Book, 0, len(books))

	for _, b := range books {
		matchTitle := strings.Contains(strings.ToLower(b.Title), search)
		matchAuthor := strings.Contains(strings.ToLower(b.Author), search)
		matchGenre := strings.Contains(strings.ToLower(b.Genre), search)

		if matchTitle || matchAuthor || matchGenre {
			result = append(result, b)
		}
	}

	return result
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	var input CreateBookInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Invalid request body")
		return
	}

	if input.Title == "" || input.Author == "" {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Title and author are required")
		return
	}

	newBook, err := h.repo.Create(userID, input)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Failed to create book")
		return
	}
	response.Success(w, http.StatusCreated, "Success Created Book", newBook)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var input UpdateBookInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Invalid request body")
		return
	}

	updated, err := h.repo.Update(userID, id, input)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			response.Error(w, http.StatusNotFound, response.CodeNotFound, "Book not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Something went wrong")
		return
	}

	response.Success(w, http.StatusOK, "Success Updated Book", updated)
}

func (h *Handler) toggleFavorite(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	updated, err := h.repo.ToggleFavorite(userID, id)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			response.Error(w, http.StatusNotFound, response.CodeNotFound, "Book not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Something went wrong")
		return
	}

	response.Success(w, http.StatusOK, "Success Toggled Favorite", updated)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	if err := h.repo.Delete(userID, id); err != nil {
		if errors.Is(err, ErrBookNotFound) {
			response.Error(w, http.StatusNotFound, response.CodeNotFound, "Book not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Something went wrong")
		return
	}

	response.Success(w, http.StatusOK, "Success Deleted Book", nil)
}
