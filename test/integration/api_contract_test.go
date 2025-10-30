//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestLoginValidationError(t *testing.T) {
	resetDatabase(t)

	server := newTestAppServer(t)
	defer server.Close()

	body, err := json.Marshal(map[string]string{})
	if err != nil {
		t.Fatalf("marshal empty payload: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.baseURL+"/api/v1/login", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.client.Do(req)
	if err != nil {
		t.Fatalf("execute login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("login validation: expected 400, got %s", resp.Status)
	}

	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["error"] != "invalid request body" {
		t.Fatalf("expected invalid request body error, got %q", payload["error"])
	}
}

func TestCreateCompanyUnauthorized(t *testing.T) {
	resetDatabase(t)

	server := newTestAppServer(t)
	defer server.Close()

	body, err := json.Marshal(map[string]any{
		"name":                "Unauthorized",
		"amount_of_employees": 10,
		"registered":          true,
		"type":                "Corporations",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.baseURL+"/api/v1/companies", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.client.Do(req)
	if err != nil {
		t.Fatalf("execute create request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized create: expected 401, got %s", resp.Status)
	}

	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["error"] != "authorization header required" {
		t.Fatalf("expected authorization header required error, got %q", payload["error"])
	}
}

func TestCreateCompanyValidationError(t *testing.T) {
	resetDatabase(t)

	server := newTestAppServer(t)
	defer server.Close()

	const (
		email    = "contract@example.com"
		password = "secret123"
	)

	seedTestUser(t, email, password)
	token := obtainAuthToken(t, server, email, password)

	body, err := json.Marshal(map[string]any{
		"name":                "ThisNameIsDefinitelyTooLong",
		"amount_of_employees": 1,
		"registered":          true,
		"type":                "Corporations",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.baseURL+"/api/v1/companies", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := server.client.Do(req)
	if err != nil {
		t.Fatalf("execute create request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("create validation: expected 400, got %s", resp.Status)
	}

	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["error"] != "name exceeds the limit of 15 characters" {
		t.Fatalf("expected name length error, got %q", payload["error"])
	}
}
