package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

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
		"component": "desolabs-app",
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) handleGetAll(w http.ResponseWriter, r *http.Request) {
	records, err := s.ds.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, "Datastore error: "+err.Error())
		return
	}

	resp := map[string]interface{}{
		"source":  "datastore",
		"records": records,
		"count":   len(records),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID: "+r.PathValue("id"))
		return
	}

	record, err := s.ds.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Datastore error: "+err.Error())
		return
	}
	if record == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Record %d not found", id))
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	err := s.ds.Health(r.Context())
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":              "error",
			"datastore_connected": false,
			"error":               err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":              "ok",
		"datastore_connected": true,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
