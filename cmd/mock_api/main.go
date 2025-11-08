package main

import (
	"fmt"
	"log"
	"mail_helper_bot/internal/pkg/mock-api/handlers"
	"net/http"
)

func main() {
	// Настройка маршрутов
	http.HandleFunc("/api/v1/private/mkdir/", handlers.MkdirHandler)
	http.HandleFunc("/api/v1/private/add", handlers.AddHandler)
	http.HandleFunc("/api/v1/private/share/", handlers.ShareHandler)
	http.HandleFunc("/api/v1/private/unshare/", handlers.UnshareHandler)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// CORS middleware (если нужно)
	handler := corsMiddleware(http.DefaultServeMux)

	port := ":8082"
	fmt.Printf("Mock API Server запущен на порту %s\n", port)
	fmt.Println("Доступные эндпоинты:")
	fmt.Println("   POST /api/v1/private/mkdir/{path}")
	fmt.Println("   POST /api/v1/private/add")
	fmt.Println("   POST /api/v1/private/share/{path}")
	fmt.Println("   POST /api/v1/private/unshare/{path}")
	fmt.Println("   GET  /health")

	log.Fatal(http.ListenAndServe(port, handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Настройка CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
