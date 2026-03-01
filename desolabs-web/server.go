package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

// Server holds the HTTP server and dependencies.
type Server struct {
	port   string
	client *AppClient
	start  time.Time
}

// NewServer creates a new Server with routing configured.
func NewServer(port string, client *AppClient) *Server {
	return &Server{
		port:   port,
		client: client,
		start:  time.Now(),
	}
}

// Start runs the HTTP server and blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /info", s.handleInfo)
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readiness", s.handleReadiness)

	httpServer := &http.Server{
		Addr:         ":" + s.port,
		Handler:      withLogging(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("desolabs-web listening on :%s", s.port)
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

// isFromCurl checks if the request comes from curl (text output).
func isFromCurl(r *http.Request) bool {
	ua := r.Header.Get("User-Agent")
	return len(ua) >= 4 && ua[:4] == "curl"
}
