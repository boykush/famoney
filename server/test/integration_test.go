package test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Disable Ryuk to avoid Docker Desktop connectivity issues on macOS.
	// Cleanup is handled by t.Cleanup instead.
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()

	// Get project root (server/test -> project root is ../../)
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("failed to get project root: %v", err)
	}

	// 1. Start PostgreSQL container
	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("expense_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() { pgContainer.Terminate(ctx) })

	expenseDBURL, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Create family_test database using lib/pq
	baseConnStr := replaceDBName(expenseDBURL, "postgres")
	db, err := sql.Open("postgres", baseConnStr)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	if _, err := db.ExecContext(ctx, "CREATE DATABASE family_test"); err != nil {
		t.Fatalf("failed to create family_test database: %v", err)
	}
	db.Close()
	familyDBURL := replaceDBName(expenseDBURL, "family_test")

	// 2. Run Atlas migrations
	expenseMigrationsDir := filepath.Join(projectRoot, "server/expense/migrations")
	atlasCmd := exec.Command("mise", "exec", "--", "atlas", "migrate", "apply",
		"--dir", fmt.Sprintf("file://%s", expenseMigrationsDir),
		"--url", expenseDBURL,
	)
	atlasCmd.Dir = projectRoot
	atlasCmd.Stdout = os.Stdout
	atlasCmd.Stderr = os.Stderr
	if err := atlasCmd.Run(); err != nil {
		t.Fatalf("failed to run expense atlas migrations: %v", err)
	}

	familyMigrationsDir := filepath.Join(projectRoot, "server/family/migrations")
	familyAtlasCmd := exec.Command("mise", "exec", "--", "atlas", "migrate", "apply",
		"--dir", fmt.Sprintf("file://%s", familyMigrationsDir),
		"--url", familyDBURL,
	)
	familyAtlasCmd.Dir = projectRoot
	familyAtlasCmd.Stdout = os.Stdout
	familyAtlasCmd.Stderr = os.Stderr
	if err := familyAtlasCmd.Run(); err != nil {
		t.Fatalf("failed to run family atlas migrations: %v", err)
	}

	// 3. Start services using free ports to avoid conflicts
	expensePort := freePort(t)
	familyPort := freePort(t)
	bffPort := freePort(t)

	// Start expense service
	expenseCmd := exec.Command(filepath.Join(projectRoot, "server/expense/bin/server"))
	expenseCmd.Env = append(os.Environ(),
		fmt.Sprintf("EXPENSE_SERVICE_PORT=%s", expensePort),
		fmt.Sprintf("DATABASE_URL=%s", expenseDBURL),
	)
	expenseCmd.Stdout = os.Stdout
	expenseCmd.Stderr = os.Stderr
	if err := expenseCmd.Start(); err != nil {
		t.Fatalf("failed to start expense service: %v", err)
	}
	t.Cleanup(func() { expenseCmd.Process.Kill(); expenseCmd.Wait() })

	// Start family service
	familyCmd := exec.Command(filepath.Join(projectRoot, "server/family/bin/server"))
	familyCmd.Env = append(os.Environ(),
		fmt.Sprintf("FAMILY_SERVICE_PORT=%s", familyPort),
		fmt.Sprintf("DATABASE_URL=%s", familyDBURL),
	)
	familyCmd.Stdout = os.Stdout
	familyCmd.Stderr = os.Stderr
	if err := familyCmd.Start(); err != nil {
		t.Fatalf("failed to start family service: %v", err)
	}
	t.Cleanup(func() { familyCmd.Process.Kill(); familyCmd.Wait() })

	// Start BFF service
	bffCmd := exec.Command(filepath.Join(projectRoot, "server/bff/bin/server"))
	bffCmd.Env = append(os.Environ(),
		fmt.Sprintf("BFF_HTTP_PORT=%s", bffPort),
		fmt.Sprintf("EXPENSE_SERVICE_ADDR=localhost:%s", expensePort),
		fmt.Sprintf("FAMILY_SERVICE_ADDR=localhost:%s", familyPort),
	)
	bffCmd.Stdout = os.Stdout
	bffCmd.Stderr = os.Stderr
	if err := bffCmd.Start(); err != nil {
		t.Fatalf("failed to start bff service: %v", err)
	}
	t.Cleanup(func() { bffCmd.Process.Kill(); bffCmd.Wait() })

	// 4. Wait for BFF health
	healthURL := fmt.Sprintf("http://localhost:%s/api/v1/expense/health", bffPort)
	waitForHealth(t, healthURL, 30*time.Second)

	// 5. Run hurl tests
	hurlCmd := exec.Command("mise", "exec", "--", "hurl",
		"--variable", fmt.Sprintf("bff_port=%s", bffPort),
		"--test",
		filepath.Join(projectRoot, "server/test/expense_health.hurl"),
		filepath.Join(projectRoot, "server/test/expense_operations.hurl"),
		filepath.Join(projectRoot, "server/test/family_health.hurl"),
	)
	hurlCmd.Dir = projectRoot
	hurlCmd.Stdout = os.Stdout
	hurlCmd.Stderr = os.Stderr
	if err := hurlCmd.Run(); err != nil {
		t.Fatalf("hurl tests failed: %v", err)
	}
}

func replaceDBName(connStr, newDB string) string {
	// Connection string format: postgres://user:pass@host:port/dbname?params
	qIdx := strings.Index(connStr, "?")
	query := ""
	base := connStr
	if qIdx >= 0 {
		query = connStr[qIdx:]
		base = connStr[:qIdx]
	}
	lastSlash := strings.LastIndex(base, "/")
	if lastSlash < 0 {
		return connStr
	}
	return base[:lastSlash+1] + newDB + query
}

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return fmt.Sprintf("%d", port)
}

func waitForHealth(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("health check at %s did not pass within %v", url, timeout)
}
