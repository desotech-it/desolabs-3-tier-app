package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Record represents a country entry from the datastore.
type Record struct {
	ID         int    `json:"id"`
	Country    string `json:"country"`
	Capital    string `json:"capital"`
	Population int    `json:"population"`
}

// DataSource defines the interface for fetching data.
// Implementations can use HTTP (current), PostgreSQL, Redis, gRPC, etc.
type DataSource interface {
	GetAll(ctx context.Context) ([]Record, error)
	GetByID(ctx context.Context, id int) (*Record, error)
	Health(ctx context.Context) error
}

// HTTPDataSource implements DataSource by calling the desolabs-data REST API.
type HTTPDataSource struct {
	baseURL string
	client  *http.Client
}

// NewHTTPDataSource creates a new HTTPDataSource pointing to the given base URL.
func NewHTTPDataSource(baseURL string) *HTTPDataSource {
	return &HTTPDataSource{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (ds *HTTPDataSource) GetAll(ctx context.Context) ([]Record, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ds.baseURL+"/data", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := ds.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call datastore: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("datastore returned %d: %s", resp.StatusCode, string(body))
	}

	var records []Record
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return records, nil
}

func (ds *HTTPDataSource) GetByID(ctx context.Context, id int) (*Record, error) {
	url := fmt.Sprintf("%s/data/%d", ds.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := ds.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call datastore: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("datastore returned %d: %s", resp.StatusCode, string(body))
	}

	var record Record
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &record, nil
}

func (ds *HTTPDataSource) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ds.baseURL+"/healthz", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := ds.client.Do(req)
	if err != nil {
		return fmt.Errorf("datastore unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("datastore unhealthy: status %d", resp.StatusCode)
	}
	return nil
}
