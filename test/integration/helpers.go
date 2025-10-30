//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ktsiligkos/xm_project/internal/platform/app"
	"github.com/ktsiligkos/xm_project/pkg/config"
)

const (
	defaultTimeout       = 5 * time.Second
	defaultUserID  int64 = 1
)

type loginResponse struct {
	Token string `json:"token"`
}

type testAppServer struct {
	app     *app.Application
	server  *httptest.Server
	client  *http.Client
	baseURL string
}

func newTestAppServer(t testing.TB) *testAppServer {
	t.Helper()

	cfg := config.Config{
		HTTPAddr:     ":0",
		MySQLDSN:     getenv("INTEGRATION_MYSQL_DSN", "xm:xmpass@tcp(127.0.0.1:3307)/xm_companies"),
		JWTSecret:    getenv("INTEGRATION_JWT_SECRET", "secret1234"),
		KafkaBrokers: parseCSV(getenv("INTEGRATION_KAFKA_BROKERS", "localhost:9094")),
		KafkaTopic:   getenv("INTEGRATION_KAFKA_TOPIC", "company-events"),
	}

	application, err := app.New(cfg)
	if err != nil {
		t.Fatalf("init application: %v", err)
	}

	server := httptest.NewServer(application.Handler())

	return &testAppServer{
		app:     application,
		server:  server,
		client:  server.Client(),
		baseURL: server.URL,
	}
}

func (s *testAppServer) Close() {
	if s.server != nil {
		s.server.Close()
	}
	if s.app != nil {
		_ = s.app.Close()
	}
}

// Reset database state for the next test run.
func resetDatabase(t testing.TB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if err := execStatements(ctx, testDB, []string{
		"DELETE FROM companies",
		"DELETE FROM users",
	}); err != nil {
		t.Fatalf("reset database: %v", err)
	}
}

func execStatements(ctx context.Context, db *sql.DB, statements []string) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	for _, stmt := range statements {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec %q: %w", stmt, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func seedTestUser(t testing.TB, email, password string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	const insertUser = `
INSERT INTO users (id, name, email, password_hash)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  password_hash = VALUES(password_hash)
`
	if _, err := testDB.ExecContext(ctx, insertUser, defaultUserID, "Integration User", email, string(hashed)); err != nil {
		t.Fatalf("insert test user: %v", err)
	}
}

func obtainAuthToken(t testing.TB, srv *testAppServer, email, password string) string {
	t.Helper()

	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal login payload: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, srv.baseURL+"/api/v1/login", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.client.Do(req)
	if err != nil {
		t.Fatalf("execute login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: expected 200 OK, got %s", resp.Status)
	}

	var login loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	if login.Token == "" {
		t.Fatalf("login response missing token")
	}

	return login.Token
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		result = []string{"localhost:9094"}
	}
	return result
}
