package main

import (
	"log"
	"net/http"

	"reading-list-api/internal/book"
)

func main() {
	repo := book.NewRepository()
	handler := book.NewHandler(repo)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

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
