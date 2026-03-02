//go:build integration

package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/kuromii5/chat-bot-auth-service/config"
	pgadapter "github.com/kuromii5/chat-bot-auth-service/internal/adapters/postgres"
)

const (
	testDBName   = "test_auth"
	testUser     = "test"
	testPassword = "test"
)

// shared across all tests in the package
var (
	testDB   *sqlx.DB
	testRepo testRepoInterface
)

// testRepoInterface combines all adapter methods for convenience.
type testRepoInterface interface {
	userRepo
	tokenRepo
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, connStr, err := startPostgres(ctx)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}

	testDB, err = sqlx.Connect("pgx", connStr)
	if err != nil {
		log.Fatalf("connect to test db: %v", err)
	}

	if err := applyMigrations(ctx, testDB); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	// create repo via adapter's New() using mapped port
	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("get container host: %v", err)
	}
	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("get mapped port: %v", err)
	}

	repo, err := pgadapter.New(&config.DatabaseConfig{
		Host:     host,
		Port:     mappedPort.Port(),
		User:     testUser,
		Password: testPassword,
		DBName:   testDBName,
		SSLMode:  "disable",
	})
	if err != nil {
		log.Fatalf("create postgres adapter: %v", err)
	}
	testRepo = repo

	code := m.Run()

	testDB.Close()
	container.Terminate(ctx)
	os.Exit(code)
}

func startPostgres(ctx context.Context) (testcontainers.Container, string, error) {
	container, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testUser),
		postgres.WithPassword(testPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, "", fmt.Errorf("run postgres container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, "", fmt.Errorf("get connection string: %w", err)
	}

	return container, connStr, nil
}

func applyMigrations(_ context.Context, db *sqlx.DB) error {
	// Create schemas manually (migrations 001, 002 use env var placeholders)
	for _, schema := range []string{"auth", "core"} {
		if _, err := db.Exec("CREATE SCHEMA IF NOT EXISTS " + schema); err != nil {
			return fmt.Errorf("create schema %s: %w", schema, err)
		}
	}


	// Apply table DDL from migration files 003, 004, 005
	migrationsDir := filepath.Join("..", "..", "..", "chat-bot-migrations", "migrations")
	files := []string{
		"003_create_auth_users_table.sql",
		"004_create_auth_refresh_tokens_table.sql",
		"005_create_core_profiles_table.sql",
	}

	for _, f := range files {
		path := filepath.Join(migrationsDir, f)
		sql, err := extractUpSQL(path)
		if err != nil {
			return fmt.Errorf("extract up sql from %s: %w", f, err)
		}
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("exec migration %s: %w", f, err)
		}
	}

	return nil
}

// extractUpSQL reads a migration file and returns everything between
// "-- +migrate Up" and "-- +migrate Down" markers.
// It also strips "-- +migrate StatementBegin/End" markers.
func extractUpSQL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := string(data)

	// Extract Up section
	upIdx := strings.Index(content, "-- +migrate Up")
	if upIdx == -1 {
		return "", fmt.Errorf("no '-- +migrate Up' marker in %s", path)
	}
	content = content[upIdx+len("-- +migrate Up"):]

	downIdx := strings.Index(content, "-- +migrate Down")
	if downIdx != -1 {
		content = content[:downIdx]
	}

	// Strip StatementBegin/End markers
	content = strings.ReplaceAll(content, "-- +migrate StatementBegin", "")
	content = strings.ReplaceAll(content, "-- +migrate StatementEnd", "")

	return strings.TrimSpace(content), nil
}

func truncateAll(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		testDB.MustExec("TRUNCATE auth.refresh_tokens, core.profiles, auth.users CASCADE")
	})
}
