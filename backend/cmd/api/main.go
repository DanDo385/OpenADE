package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"openade/internal/db"
	"openade/internal/handlers"
	"openade/internal/services"
)

func main() {
	// --- Configuration from environment ---
	port := envOr("OPENADE_PORT", "8080")
	dbPath := envOr("OPENADE_DB_PATH", "./openade.db")

	// --- Database ---
	log.Printf("Opening database at %s", dbPath)
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()
	log.Println("Database ready")

	// --- Services ---
	convSvc := services.NewConversationService(database)
	taskSvc := services.NewTaskService(database)
	runSvc := services.NewRunService(database)
	memSvc := services.NewMemoryService(database)
	provSvc := services.NewProviderService(database)

	// --- HTTP Server ---
	srv := handlers.NewServer(convSvc, taskSvc, runSvc, memSvc, provSvc)

	r := chi.NewRouter()

	// Middleware stack
	r.Use(handlers.RecoveryMiddleware)
	r.Use(handlers.LoggingMiddleware)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://localhost:1420", "tauri://localhost"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Register routes
	srv.RegisterRoutes(r)

	// --- Start server ---
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // Long timeout for SSE streaming
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown on SIGTERM / SIGINT
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("OpenADE backend listening on :%s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-done
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
