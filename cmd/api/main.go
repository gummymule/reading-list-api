package main

import (
	"log"
	"net/http"

	"reading-list-api/internal/book"
	"reading-list-api/internal/database"
	"reading-list-api/internal/goal"
	"reading-list-api/internal/user"
)

func main() {
	db, err := database.New("reading_list.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := user.NewRepository(db)
	userHandler := user.NewHandler(userRepo)

	bookRepo := book.NewRepository(db)
	bookHandler := book.NewHandler(bookRepo)

	goalRepo := goal.NewRepository(db)
	goalHandler := goal.NewHandler(goalRepo, func(userID string) int {
		return bookRepo.CountByStatus(userID, book.StatusRead)
	})

	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)
	bookHandler.RegisterRoutes(mux)
	goalHandler.RegisterRoutes(mux)

	corsHandler := withCORS(mux)

	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", corsHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
