package goal

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"reading-list-api/internal/auth"
	"reading-list-api/internal/response"
)

type countReadBooksFunc func(userID string) int

type Handler struct {
	repo           *Repository
	countReadBooks countReadBooksFunc
}

func NewHandler(repo *Repository, countReadBooks countReadBooksFunc) *Handler {
	return &Handler{repo: repo, countReadBooks: countReadBooks}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/reading-goal", auth.RequireAuth(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /api/reading-goal", auth.RequireAuth(http.HandlerFunc(h.update)))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	year := time.Now().Year()
	if yearParam := r.URL.Query().Get("year"); yearParam != "" {
		if parsed, err := strconv.Atoi(yearParam); err == nil {
			year = parsed
		}
	}

	goal := ReadingGoal{
		Year:    year,
		Target:  h.repo.GetTarget(userID, year),
		Current: h.countReadBooks(userID),
	}

	response.Success(w, http.StatusOK, "Success Get Reading Goal", goal)

}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	var input UpdateGoalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Invalid request body")
		return
	}

	if input.Target < 0 {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Target must be zero or greater")
		return
	}

	year := input.Year
	if year == 0 {
		year = time.Now().Year()
	}

	if err := h.repo.SetTarget(userID, year, input.Target); err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Failed to update reading goal")
		return
	}

	goal := ReadingGoal{
		Year:    year,
		Target:  input.Target,
		Current: h.countReadBooks(userID),
	}

	response.Success(w, http.StatusOK, "Success Updated Reading Goal", goal)
}
