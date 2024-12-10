package main

import (
	"log"
	"net/http"
	"github.com/joho/godotenv"
    "olobillz-server/internal/app"
    "os"
)

func main() {
	mux := http.NewServeMux()

	// Sample health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is running, there is a joy in every hello."))
	})

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
