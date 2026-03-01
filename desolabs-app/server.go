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
	ds    DataSource
	start time.Time
}

// NewServer creates a new Server with routing configured.
func NewServer(port string, ds DataSource) *Server {
	return &Server{
		port:  port,
		ds:    ds,
		start: time.Now(),
	}
}

// Start runs the HTTP server and blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("GET /api/info", s.handleInfo)
	mux.HandleFunc("GET /api/data", s.handleGetAll)
	mux.HandleFunc("GET /api/data/{id}", s.handleGetByID)

	// Health
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readiness", s.handleReadiness)

	httpServer := &http.Server{
		Addr:         ":" + s.port,
		Handler:      withLogging(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("desolabs-app listening on :%s", s.port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
}
