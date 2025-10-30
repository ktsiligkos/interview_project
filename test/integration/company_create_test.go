//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

type companyResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       *string `json:"description"`
	AmountOfEmployees int     `json:"amount_of_employees"`
	Registered        bool    `json:"registered"`
	Type              string  `json:"type"`
}

func TestCreateCompanyIntegration(t *testing.T) {
	// resetDatabase(t)

	const (
		email    = "john_doe@example.com"
		password = "12345678"
	)

	server := newTestAppServer(t)
	defer server.Close()

	seedTestUser(t, email, password)
	token := obtainAuthToken(t, server, email, password)

	companyName := fmt.Sprintf("Comp%d", time.Now().UnixNano()%100000)
	description := "Integration created company"

	payload := map[string]any{
		"name":                companyName,
		"description":         description,
		"amount_of_employees": 42,
		"registered":          true,
		"type":                "Corporations",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal create payload: %v", err)
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

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create company: expected 201 Created, got %s", resp.Status)
	}

	var created companyResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	if created.ID == "" {
		t.Fatalf("expected company id in response")
	}
	if created.Name != companyName {
		t.Fatalf("expected company name %q, got %q", companyName, created.Name)
	}

	assertCompanyInDB(t, created.ID, companyName, description, 42, true, "Corporations")

	fetched := fetchCompanyByID(t, server, created.ID)
	if fetched.Name != companyName {
		t.Fatalf("expected fetched name %q, got %q", companyName, fetched.Name)
	}
	if fetched.Description == nil || *fetched.Description != description {
		t.Fatalf("unexpected fetched description: %#v", fetched.Description)
	}
}

func assertCompanyInDB(t testing.TB, id, expectedName, expectedDescription string, expectedEmployees int, expectedRegistered bool, expectedType string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	const query = `
SELECT name, description, amount_of_employees, registered, type
FROM companies
WHERE id = ?
`

	var (
		name        string
		description sql.NullString
		employees   int
		registered  bool
		ctype       string
	)

	if err := testDB.QueryRowContext(ctx, query, id).Scan(&name, &description, &employees, &registered, &ctype); err != nil {
		t.Fatalf("query company from db: %v", err)
	}

	if name != expectedName {
		t.Fatalf("db name mismatch: want %q got %q", expectedName, name)
	}

	if !description.Valid || description.String != expectedDescription {
		t.Fatalf("db description mismatch: want %q got %q", expectedDescription, description.String)
	}

	if employees != expectedEmployees {
		t.Fatalf("db employees mismatch: want %d got %d", expectedEmployees, employees)
	}

	if registered != expectedRegistered {
		t.Fatalf("db registered mismatch: want %v got %v", expectedRegistered, registered)
	}

	if ctype != expectedType {
		t.Fatalf("db type mismatch: want %q got %q", expectedType, ctype)
	}
}

func fetchCompanyByID(t testing.TB, server *testAppServer, id string) companyResponse {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.baseURL+"/api/v1/companies/"+id, nil)
	if err != nil {
		t.Fatalf("build get request: %v", err)
	}

	resp, err := server.client.Do(req)
	if err != nil {
		t.Fatalf("execute get request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get company: expected 200 OK, got %s", resp.Status)
	}

	var result companyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode get response: %v", err)
	}

	return result
}
