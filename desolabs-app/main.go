package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("desolabs-app %s (commit: %s, built: %s)", version, commit, date)

	dataURL := getEnv("DATA_URL", "http://desolabs-data:5000")
	port := getEnv("PORT", "3001")

	ds := NewHTTPDataSource(dataURL)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := NewServer(port, ds)
	srv.Start(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
