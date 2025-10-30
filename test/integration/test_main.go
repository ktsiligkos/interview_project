//go:build integration

// Package integration exercises the running application stack via HTTP requests.
// Start the support services (e.g. `docker compose up -d`) before running
// `go test -tags=integration ./test/integration/...`.
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	dsn := getenv("INTEGRATION_MYSQL_DSN", "xm:xmpass@tcp(127.0.0.1:3307)/xm_companies")

	var err error
	testDB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open mysql connection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := waitForDatabase(ctx); err != nil {
		log.Fatalf("wait for database: %v", err)
	}

	code := m.Run()

	if testDB != nil {
		_ = testDB.Close()
	}

	os.Exit(code)
}

func waitForDatabase(ctx context.Context) error {
	const (
		maxAttempts = 30
		delay       = 2 * time.Second
	)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := testDB.PingContext(ctx); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("database not ready: %w", ctx.Err())
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("database not reachable after %d attempts", maxAttempts)
}
