//go:build integration

package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/Discard3700/discord-festival-bot/internal/db"
)

func testDatabaseURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test")
	}
	return url
}

func TestConnect(t *testing.T) {
	url := testDatabaseURL(t)
	pool, err := db.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer pool.Close()
}

func TestMigrate(t *testing.T) {
	url := testDatabaseURL(t)
	if err := db.Migrate(url); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	url := testDatabaseURL(t)
	if err := db.Migrate(url); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	// second run must be a no-op, not an error
	if err := db.Migrate(url); err != nil {
		t.Fatalf("second Migrate (should be no-op): %v", err)
	}
}
