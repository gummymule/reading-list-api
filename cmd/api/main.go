package main

import (
	"log"
	"net/http"

	"reading-list-api/internal/book"
	"reading-list-api/internal/database"
	"reading-list-api/internal/goal"
)

func main() {
	db, err := database.New("reading_list.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	bookRepo := book.NewRepository(db)
	bookHandler := book.NewHandler(bookRepo)

	goalRepo := goal.NewRepository(db)
	goalHandler := goal.NewHandler(goalRepo, func() int {
		return bookRepo.CountByStatus(book.StatusRead)
	})

	mux := http.NewServeMux()
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
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
