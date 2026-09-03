package db

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/rodrigueghenda/jobira/internal/config"
)

func TestNewPostgres_MissingDatabaseURL(
	t *testing.T,
) {
	cfg := config.Config{}

	database, err := NewPostgres(cfg)

	if err == nil {
		t.Fatal(
			"expected missing DATABASE_URL error",
		)
	}

	if database != nil {
		t.Fatal(
			"expected nil database",
		)
	}

	if err.Error() != "DATABASE_URL is required" {
		t.Fatalf(
			"unexpected error %q",
			err.Error(),
		)
	}
}

func TestNewPostgres_InvalidDatabaseURL(
	t *testing.T,
) {
	cfg := config.Config{
		DatabaseURL: "://invalid-database-url",
	}

	database, err := NewPostgres(cfg)

	if err == nil {
		if database != nil {
			database.Close()
		}

		t.Fatal(
			"expected invalid database URL error",
		)
	}

	if database != nil {
		database.Close()

		t.Fatal(
			"expected nil database on ping failure",
		)
	}

	if !strings.Contains(
		err.Error(),
		"ping postgres",
	) {
		t.Fatalf(
			"expected wrapped postgres ping error, got %q",
			err.Error(),
		)
	}
}

func TestNewPostgres_InvalidConnectionInput(
	t *testing.T,
) {
	cfg := config.Config{
		DatabaseURL: "not-a-valid-postgres-url",
	}

	database, err := NewPostgres(cfg)

	if database != nil {
		database.Close()
	}

	if err == nil {
		t.Fatal(
			"expected connection error",
		)
	}

	if strings.TrimSpace(
		err.Error(),
	) == "" {
		t.Fatal(
			"expected non-empty error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"ping postgres",
	) {
		t.Fatalf(
			"expected wrapped postgres error, got %q",
			err.Error(),
		)
	}
}

func TestConfigurePool(
	t *testing.T,
) {
	database, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"sqlmock.New: %v",
			err,
		)
	}
	defer database.Close()

	configurePool(database)

	stats := database.Stats()

	if stats.MaxOpenConnections != maxOpenConnections {
		t.Fatalf(
			"expected max open connections %d, got %d",
			maxOpenConnections,
			stats.MaxOpenConnections,
		)
	}
}
