//go:build integration

package lineup_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Discard3700/discord-festival-bot/internal/db"
	"github.com/Discard3700/discord-festival-bot/internal/lineup"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	if err := db.Migrate(url); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := db.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestUpsertFestivalAndArtists(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	loc, _ := time.LoadLocation("America/Chicago")

	fest := lineup.Festival{
		Name:     "Integration Test Festival",
		Slug:     "integration-test-2025",
		StartsAt: time.Date(2025, 7, 11, 0, 0, 0, 0, loc),
		EndsAt:   time.Date(2025, 7, 13, 23, 59, 59, 0, loc),
	}
	festID, err := lineup.UpsertFestival(ctx, pool, fest)
	if err != nil {
		t.Fatalf("UpsertFestival: %v", err)
	}
	if festID == 0 {
		t.Fatal("UpsertFestival returned 0 ID")
	}

	// Idempotent second upsert
	festID2, err := lineup.UpsertFestival(ctx, pool, fest)
	if err != nil {
		t.Fatalf("UpsertFestival (second): %v", err)
	}
	if festID2 != festID {
		t.Errorf("second UpsertFestival returned different ID: %d vs %d", festID2, festID)
	}

	friday := time.Date(2025, 7, 11, 21, 0, 0, 0, loc)
	artist := lineup.Artist{
		FestivalID: festID,
		Name:       "Integration Test Artist",
		Stage:      "Test Stage",
		StartsAt:   friday,
		EndsAt:     friday.Add(90 * time.Minute),
	}

	if err := lineup.Upsert(ctx, pool, artist); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	// Idempotent
	if err := lineup.Upsert(ctx, pool, artist); err != nil {
		t.Fatalf("Upsert (second): %v", err)
	}

	results, err := lineup.ByArtist(ctx, pool, festID, "Integration Test Artist")
	if err != nil {
		t.Fatalf("ByArtist: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("ByArtist returned no results after upsert")
	}
	if results[0].Name != artist.Name {
		t.Errorf("Name = %q, want %q", results[0].Name, artist.Name)
	}

	byDay, err := lineup.ByDay(ctx, pool, festID, time.Friday, "", loc)
	if err != nil {
		t.Fatalf("ByDay: %v", err)
	}
	found := false
	for _, a := range byDay {
		if a.Name == artist.Name {
			found = true
		}
	}
	if !found {
		t.Error("ByDay did not return the upserted artist")
	}

	f, err := lineup.FestivalBySlug(ctx, pool, "integration-test-2025")
	if err != nil {
		t.Fatalf("FestivalBySlug: %v", err)
	}
	if f.ID != festID {
		t.Errorf("FestivalBySlug ID = %d, want %d", f.ID, festID)
	}
}
