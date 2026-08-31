package user

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
	mux.HandleFunc("POST /api/auth/register", h.register)
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.HandleFunc("POST /api/auth/logout", h.logout)
	mux.HandleFunc("POST /api/auth/forgot-password", h.forgotPassword)
	mux.HandleFunc("POST /api/auth/reset-password", h.resetPassword)
	mux.Handle("GET /api/auth/profile", auth.RequireAuth(http.HandlerFunc(h.getProfile)))
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Invalid request body")
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if input.Email == "" || input.Password == "" || input.Name == "" {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Email, password, and name are required")
		return
	}

	if len(input.Password) < 8 {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Failed to process password")
		return
	}

	newUser, err := h.repo.Create(input.Email, hash, input.Name)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Email is already registered")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Failed to create account")
		return
	}

	if err := h.issueSession(w, newUser.ID); err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Failed to create session")
		return
	}

	response.Success(w, http.StatusCreated, "Success Registered", newUser)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Invalid request body")
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	existingUser, passwordHash, err := h.repo.GetByEmailWithHash(input.Email)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, response.CodeUnauthorized, "Invalid email or password")
		return
	}

	if !auth.CheckPassword(input.Password, passwordHash) {
		response.Error(w, http.StatusUnauthorized, response.CodeUnauthorized, "Invalid email or password")
		return
	}

	if err := h.issueSession(w, existingUser.ID); err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Failed to create session")
		return
	}

	response.Success(w, http.StatusOK, "Success Login", existingUser)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
	response.Success(w, http.StatusOK, "Success Logout", nil)
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, response.CodeUnauthorized, "Not authenticated")
		return
	}

	currentUser, err := h.repo.GetByID(userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, response.CodeNotFound, "User not found")
		return
	}
	response.Success(w, http.StatusOK, "Success Get Profile", currentUser)
}

func (h *Handler) issueSession(w http.ResponseWriter, userID string) error {
	token, err := auth.GenerateToken(userID)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
	return nil
}

func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var input ForgotPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Invalid request body")
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	otp, err := h.repo.CreatePasswordReset(input.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Pesan generic — tidak membocorkan apakah email terdaftar atau tidak
			response.Success(w, http.StatusOK, "If the email exists, a reset code has been generated", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Failed to process request")
		return
	}

	log.Printf("[DEV ONLY] Password reset OTP for %s: %s", input.Email, otp)

	// PENTING: OTP dikembalikan langsung di response karena belum ada email service.
	// Ini TIDAK AMAN untuk production — siapapun yang bisa lihat/intercept response ini
	// bisa reset password orang lain. Nanti WAJIB diganti kirim via email sebelum deploy sungguhan.
	response.Success(w, http.StatusOK, "Reset code generated (dev mode)", map[string]string{
		"otp": otp,
	})
}

func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var input ResetPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Invalid request body")
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if len(input.NewPassword) < 8 {
		response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(input.NewPassword)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Failed to process password")
		return
	}

	if err := h.repo.ResetPassword(input.Email, input.OTP, hash); err != nil {
		if errors.Is(err, ErrInvalidOTP) {
			response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "Invalid or expired code")
			return
		}
		response.Error(w, http.StatusInternalServerError, response.CodeInternalError, "Failed to reset password")
		return
	}

	response.Success(w, http.StatusOK, "Success Reset Password", nil)
}
