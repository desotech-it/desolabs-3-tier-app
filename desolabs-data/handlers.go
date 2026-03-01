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

func (s *Server) handleGetAll(w http.ResponseWriter, r *http.Request) {
	records, err := s.store.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch records: "+err.Error())
		return
	}
	if records == nil {
		records = []Record{}
	}
	writeJSON(w, http.StatusOK, records)
}

func (s *Server) handleGetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID: "+r.PathValue("id"))
		return
	}

	record, err := s.store.GetByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch record: "+err.Error())
		return
	}
	if record == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Record %d not found", id))
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var rec Record
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	if rec.Country == "" || rec.Capital == "" {
		writeError(w, http.StatusBadRequest, "country and capital are required")
		return
	}

	created, err := s.store.Create(rec)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create record: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID: "+r.PathValue("id"))
		return
	}

	var rec Record
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	updated, err := s.store.Update(id, rec)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update record: "+err.Error())
		return
	}
	if updated == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Record %d not found", id))
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID: "+r.PathValue("id"))
		return
	}

	if err := s.store.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Record %d deleted", id)})
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	count, _ := s.store.Count()

	info := map[string]interface{}{
		"hostname":  hostname,
		"ip":        os.Getenv("POD_IP"),
		"node":      os.Getenv("NODE_NAME"),
		"namespace": os.Getenv("POD_NAMESPACE"),
		"pod":       os.Getenv("POD_NAME"),
		"records":   count,
		"uptime":    time.Since(s.start).String(),
		"version":   version,
	}
	writeJSON(w, http.StatusOK, info)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
