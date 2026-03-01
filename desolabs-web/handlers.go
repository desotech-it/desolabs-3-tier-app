package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

var indexTemplate = template.Must(template.ParseFS(templateFS, "templates/index.html"))

// PageData holds all data passed to the HTML template.
type PageData struct {
	// Frontend pod info
	Hostname  string
	IP        string
	Node      string
	Namespace string
	Pod       string
	Version   string

	// Backend info
	Backend      *BackendInfo
	BackendError string

	// Data
	Data      *DataResponse
	DataError string
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	hostname, _ := os.Hostname()
	data := PageData{
		Hostname:  hostname,
		IP:        os.Getenv("POD_IP"),
		Node:      os.Getenv("NODE_NAME"),
		Namespace: os.Getenv("POD_NAMESPACE"),
		Pod:       os.Getenv("POD_NAME"),
		Version:   version,
	}

	// Fetch backend info
	backendInfo, err := s.client.GetInfo(r.Context())
	if err != nil {
		data.BackendError = err.Error()
	} else {
		data.Backend = backendInfo
	}

	// Fetch data
	dataResp, err := s.client.GetData(r.Context())
	if err != nil {
		data.DataError = err.Error()
	} else {
		data.Data = dataResp
	}

	if isFromCurl(r) {
		s.renderText(w, data)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTemplate.Execute(w, data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) renderText(w http.ResponseWriter, data PageData) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════════╗\n")
	sb.WriteString("║              DesoLabs 3-Tier App                ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════╝\n\n")

	sb.WriteString("── Frontend Pod Info ──────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("  Hostname:   %s\n", data.Hostname))
	sb.WriteString(fmt.Sprintf("  IP:         %s\n", data.IP))
	sb.WriteString(fmt.Sprintf("  Node:       %s\n", data.Node))
	sb.WriteString(fmt.Sprintf("  Namespace:  %s\n", data.Namespace))
	sb.WriteString(fmt.Sprintf("  Version:    %s\n\n", data.Version))

	if data.BackendError != "" {
		sb.WriteString("── Backend ───────────────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("  ERROR: %s\n\n", data.BackendError))
	} else if data.Backend != nil {
		sb.WriteString("── Backend Pod Info ───────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("  Hostname:   %s\n", data.Backend.Hostname))
		sb.WriteString(fmt.Sprintf("  IP:         %s\n", data.Backend.IP))
		sb.WriteString(fmt.Sprintf("  Node:       %s\n", data.Backend.Node))
		sb.WriteString(fmt.Sprintf("  Version:    %s\n\n", data.Backend.Version))
	}

	if data.DataError != "" {
		sb.WriteString("── Data ──────────────────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("  ERROR: %s\n\n", data.DataError))
	} else if data.Data != nil && len(data.Data.Records) > 0 {
		sb.WriteString("── Countries ─────────────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("  %-4s %-20s %-15s %s\n", "ID", "Country", "Capital", "Population"))
		sb.WriteString("  " + strings.Repeat("─", 55) + "\n")
		for _, r := range data.Data.Records {
			sb.WriteString(fmt.Sprintf("  %-4d %-20s %-15s %d\n", r.ID, r.Country, r.Capital, r.Population))
		}
		sb.WriteString(fmt.Sprintf("\n  Source: %s | Total: %d records\n", data.Data.Source, data.Data.Count))
	}

	fmt.Fprint(w, sb.String())
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()

	info := map[string]interface{}{
		"hostname":  hostname,
		"ip":        os.Getenv("POD_IP"),
		"node":      os.Getenv("NODE_NAME"),
		"namespace": os.Getenv("POD_NAMESPACE"),
		"pod":       os.Getenv("POD_NAME"),
		"uptime":    time.Since(s.start).String(),
		"version":   version,
		"component": "desolabs-web",
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	err := s.client.Health(r.Context())
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":            "error",
			"backend_connected": false,
			"error":             err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":            "ok",
		"backend_connected": true,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}
