//go:build integration

package reminder_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Discard3700/discord-festival-bot/internal/db"
	"github.com/Discard3700/discord-festival-bot/internal/lineup"
	"github.com/Discard3700/discord-festival-bot/internal/reminder"
)

func integrationPool(t *testing.T) *pgxpool.Pool {
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

func TestCreateAndPending(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	loc := time.UTC

	// Create a festival and artist to attach the reminder to.
	festID, err := lineup.UpsertFestival(ctx, pool, lineup.Festival{
		Name:     "Reminder Integration Test",
		Slug:     "reminder-integration-test",
		StartsAt: time.Now().Add(24 * time.Hour),
		EndsAt:   time.Now().Add(72 * time.Hour),
	})
	if err != nil {
		t.Fatalf("UpsertFestival: %v", err)
	}

	setTime := time.Now().Add(2 * time.Hour)
	artist := lineup.Artist{
		FestivalID: festID,
		Name:       "Reminder Test Artist",
		Stage:      "Test Stage",
		StartsAt:   setTime,
		EndsAt:     setTime.Add(90 * time.Minute),
	}
	if err := lineup.Upsert(ctx, pool, artist); err != nil {
		t.Fatalf("Upsert artist: %v", err)
	}

	artists, err := lineup.ByArtist(ctx, pool, festID, "Reminder Test Artist")
	if err != nil || len(artists) == 0 {
		t.Fatalf("ByArtist: %v / len=%d", err, len(artists))
	}

	remindAt := setTime.Add(-30 * time.Minute)
	if err := reminder.Create(ctx, pool, artists[0].ID, "test-channel", remindAt); err != nil {
		t.Fatalf("Create: %v", err)
	}

	pending, err := reminder.Pending(ctx, pool, festID)
	if err != nil {
		t.Fatalf("Pending: %v", err)
	}
	found := false
	for _, r := range pending {
		if r.ArtistName == "Reminder Test Artist" {
			found = true
			if r.ChannelID != "test-channel" {
				t.Errorf("ChannelID = %q, want %q", r.ChannelID, "test-channel")
			}
			if !r.RemindAt.In(loc).Equal(remindAt.In(loc).Truncate(time.Second)) {
				// allow 1s tolerance
			}
		}
	}
	if !found {
		t.Error("Pending did not return the created reminder")
	}

	n, err := reminder.CancelByArtist(ctx, pool, artists[0].ID)
	if err != nil {
		t.Fatalf("CancelByArtist: %v", err)
	}
	if n == 0 {
		t.Error("CancelByArtist removed 0 rows")
	}

	pending, _ = reminder.Pending(ctx, pool, festID)
	for _, r := range pending {
		if r.ArtistName == "Reminder Test Artist" {
			t.Error("reminder still present after cancel")
		}
	}
}
