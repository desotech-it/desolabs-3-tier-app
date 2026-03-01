package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BackendInfo holds info about the backend pod.
type BackendInfo struct {
	Hostname  string `json:"hostname"`
	IP        string `json:"ip"`
	Node      string `json:"node"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Uptime    string `json:"uptime"`
	Version   string `json:"version"`
	Component string `json:"component"`
}

// DataResponse holds the data returned by the backend.
type DataResponse struct {
	Source  string   `json:"source"`
	Records []Record `json:"records"`
	Count   int      `json:"count"`
}

// Record represents a country entry.
type Record struct {
	ID         int    `json:"id"`
	Country    string `json:"country"`
	Capital    string `json:"capital"`
	Population int    `json:"population"`
}

// AppClient calls the desolabs-app backend.
type AppClient struct {
	baseURL string
	client  *http.Client
}

// NewAppClient creates a new client for the backend API.
func NewAppClient(baseURL string) *AppClient {
	return &AppClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetInfo fetches backend pod info.
func (c *AppClient) GetInfo(ctx context.Context) (*BackendInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/info", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call backend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("backend returned %d: %s", resp.StatusCode, string(body))
	}

	var info BackendInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &info, nil
}

// GetData fetches country data from the backend.
func (c *AppClient) GetData(ctx context.Context) (*DataResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/data", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call backend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("backend returned %d: %s", resp.StatusCode, string(body))
	}

	var data DataResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &data, nil
}

// Health checks if the backend is reachable.
func (c *AppClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/healthz", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("backend unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backend unhealthy: status %d", resp.StatusCode)
	}
	return nil
}
