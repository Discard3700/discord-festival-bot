package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/Discard3700/discord-festival-bot/internal/db"
	"github.com/Discard3700/discord-festival-bot/internal/lineup"
)

type importFile struct {
	Festival struct {
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		StartsAt string `json:"starts_at"`
		EndsAt   string `json:"ends_at"`
	} `json:"festival"`
	Artists []struct {
		Name     string `json:"name"`
		Stage    string `json:"stage"`
		StartsAt string `json:"starts_at"`
		EndsAt   string `json:"ends_at"`
	} `json:"artists"`
}

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: import <lineup.json>")
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	if err := db.Migrate(dbURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("read file: %v", err)
	}

	var f importFile
	if err := json.Unmarshal(data, &f); err != nil {
		log.Fatalf("parse json: %v", err)
	}

	festStart, err := time.Parse(time.RFC3339, f.Festival.StartsAt)
	if err != nil {
		log.Fatalf("festival starts_at: %v", err)
	}
	festEnd, err := time.Parse(time.RFC3339, f.Festival.EndsAt)
	if err != nil {
		log.Fatalf("festival ends_at: %v", err)
	}

	festID, err := lineup.UpsertFestival(ctx, pool, lineup.Festival{
		Name:     f.Festival.Name,
		Slug:     f.Festival.Slug,
		StartsAt: festStart,
		EndsAt:   festEnd,
	})
	if err != nil {
		log.Fatalf("upsert festival: %v", err)
	}
	log.Printf("festival %q (id=%d) ready", f.Festival.Slug, festID)

	ok, fail := 0, 0
	for _, r := range f.Artists {
		starts, err := time.Parse(time.RFC3339, r.StartsAt)
		if err != nil {
			log.Printf("skip %q: bad starts_at %q: %v", r.Name, r.StartsAt, err)
			fail++
			continue
		}
		ends, err := time.Parse(time.RFC3339, r.EndsAt)
		if err != nil {
			log.Printf("skip %q: bad ends_at %q: %v", r.Name, r.EndsAt, err)
			fail++
			continue
		}
		if err := lineup.Upsert(ctx, pool, lineup.Artist{
			FestivalID: festID,
			Name:       r.Name,
			Stage:      r.Stage,
			StartsAt:   starts,
			EndsAt:     ends,
		}); err != nil {
			log.Printf("upsert %q: %v", r.Name, err)
			fail++
			continue
		}
		ok++
	}
	log.Printf("import complete: %d upserted, %d failed", ok, fail)
	if fail > 0 {
		os.Exit(1)
	}
}
