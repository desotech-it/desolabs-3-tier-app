package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

// Server holds the HTTP server and dependencies.
type Server struct {
	port  string
	store Store
	start time.Time
}

// NewServer creates a new Server with routing configured.
func NewServer(port string, store Store) *Server {
	return &Server{
		port:  port,
		store: store,
		start: time.Now(),
	}
}

// Start runs the HTTP server and blocks until the context is cancelled.
// Implements graceful shutdown.
func (s *Server) Start(ctx context.Context) {
	mux := http.NewServeMux()

	// CRUD endpoints
	mux.HandleFunc("GET /data", s.handleGetAll)
	mux.HandleFunc("GET /data/{id}", s.handleGetByID)
	mux.HandleFunc("POST /data", s.handleCreate)
	mux.HandleFunc("PUT /data/{id}", s.handleUpdate)
	mux.HandleFunc("DELETE /data/{id}", s.handleDelete)

	// Health and info
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /info", s.handleInfo)

	httpServer := &http.Server{
		Addr:         ":" + s.port,
		Handler:      withLogging(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("desolabs-data listening on :%s", s.port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

// withLogging wraps an http.Handler with request logging.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
}
