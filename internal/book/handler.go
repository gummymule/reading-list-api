package book

import (
	"encoding/json"
	"errors"
	"net/http"

	"reading-list-api/internal/response"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/books", h.list)
	mux.HandleFunc("POST /api/books", h.create)
	mux.HandleFunc("PATCH /api/books/{id}", h.update)
	mux.HandleFunc("PATCH /api/books/{id}/favorite", h.toggleFavorite)
	mux.HandleFunc("DELETE /api/books/{id}", h.delete)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	favorite := r.URL.Query().Get("favorite")

	var books []Book
	switch {
	case favorite == "true":
		books = h.repo.GetFavorites()
	case status != "":
		books = h.repo.GetByStatus(ReadingStatus(status))
	default:
		books = h.repo.GetAll()
	}

	response.Success(w, http.StatusOK, "Success Get Books", books)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateBookInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Invalid request body")
		return
	}

	if input.Title == "" || input.Author == "" {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Title and author are required")
		return
	}

	newBook := h.repo.Create(input)
	response.Success(w, http.StatusCreated, "Success Created Book", newBook)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input UpdateBookInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Invalid request body")
		return
	}

	updated, err := h.repo.Update(id, input)
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
	id := r.PathValue("id")

	updated, err := h.repo.ToggleFavorite(id)
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
	id := r.PathValue("id")

	if err := h.repo.Delete(id); err != nil {
		if errors.Is(err, ErrBookNotFound) {
			response.Error(w, http.StatusNotFound, response.CodeNotFound, "Book not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Something went wrong")
		return
	}

	response.Success(w, http.StatusOK, "Success Deleted Book", nil)
}
