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
	log.Printf("desolabs-web %s (commit: %s, built: %s)", version, commit, date)

	appURL := getEnv("APP_URL", "http://desolabs-app:3001")
	port := getEnv("PORT", "80")

	client := NewAppClient(appURL)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := NewServer(port, client)
	srv.Start(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
