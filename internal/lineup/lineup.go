package lineup

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Festival struct {
	ID       int
	Name     string
	Slug     string
	StartsAt time.Time
	EndsAt   time.Time
}

type Artist struct {
	ID         int
	FestivalID int
	Name       string
	Stage      string
	StartsAt   time.Time
	EndsAt     time.Time
}

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "sun": time.Sunday,
	"monday": time.Monday, "mon": time.Monday,
	"tuesday": time.Tuesday, "tue": time.Tuesday,
	"wednesday": time.Wednesday, "wed": time.Wednesday,
	"thursday": time.Thursday, "thu": time.Thursday,
	"friday": time.Friday, "fri": time.Friday,
	"saturday": time.Saturday, "sat": time.Saturday,
}

// ParseWeekday converts a user-supplied string ("Friday", "fri") to time.Weekday.
func ParseWeekday(s string) (time.Weekday, bool) {
	d, ok := weekdays[strings.ToLower(strings.TrimSpace(s))]
	return d, ok
}

// UpsertFestival inserts or updates a festival by slug and returns its ID.
func UpsertFestival(ctx context.Context, pool *pgxpool.Pool, f Festival) (int, error) {
	var id int
	err := pool.QueryRow(ctx, `
		INSERT INTO festivals (name, slug, starts_at, ends_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (slug) DO UPDATE
		  SET name = EXCLUDED.name, starts_at = EXCLUDED.starts_at, ends_at = EXCLUDED.ends_at
		RETURNING id`,
		f.Name, f.Slug, f.StartsAt, f.EndsAt,
	).Scan(&id)
	return id, err
}

// DefaultFestival returns the currently active festival, then the next upcoming,
// then the most recent past one. Returns an error if no festivals exist.
func DefaultFestival(ctx context.Context, pool *pgxpool.Pool) (*Festival, error) {
	for _, q := range []string{
		`SELECT id, name, slug, starts_at, ends_at FROM festivals WHERE NOW() BETWEEN starts_at AND ends_at ORDER BY starts_at LIMIT 1`,
		`SELECT id, name, slug, starts_at, ends_at FROM festivals WHERE starts_at > NOW() ORDER BY starts_at LIMIT 1`,
		`SELECT id, name, slug, starts_at, ends_at FROM festivals WHERE ends_at < NOW() ORDER BY ends_at DESC LIMIT 1`,
	} {
		var f Festival
		err := pool.QueryRow(ctx, q).Scan(&f.ID, &f.Name, &f.Slug, &f.StartsAt, &f.EndsAt)
		if err == nil {
			return &f, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("festival query: %w", err)
		}
	}
	return nil, fmt.Errorf("no festivals found — run 'make import' first")
}

// FestivalBySlug looks up a festival by its short slug.
func FestivalBySlug(ctx context.Context, pool *pgxpool.Pool, slug string) (*Festival, error) {
	var f Festival
	err := pool.QueryRow(ctx,
		`SELECT id, name, slug, starts_at, ends_at FROM festivals WHERE slug = $1`, slug,
	).Scan(&f.ID, &f.Name, &f.Slug, &f.StartsAt, &f.EndsAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("festival %q not found", slug)
	}
	return &f, err
}

const artistCols = `SELECT id, festival_id, name, stage, starts_at, ends_at FROM artists`

func scanArtist(row interface{ Scan(...any) error }) (Artist, error) {
	var a Artist
	return a, row.Scan(&a.ID, &a.FestivalID, &a.Name, &a.Stage, &a.StartsAt, &a.EndsAt)
}

// NowPlaying returns sets currently in progress for the given festival.
func NowPlaying(ctx context.Context, pool *pgxpool.Pool, festivalID int) ([]Artist, error) {
	rows, err := pool.Query(ctx,
		artistCols+` WHERE festival_id = $1 AND NOW() BETWEEN starts_at AND ends_at ORDER BY stage, starts_at`,
		festivalID,
	)
	if err != nil {
		return nil, fmt.Errorf("now playing: %w", err)
	}
	defer rows.Close()
	return collectArtists(rows)
}

// NextPlaying returns sets starting within the next hour for the given festival.
func NextPlaying(ctx context.Context, pool *pgxpool.Pool, festivalID int) ([]Artist, error) {
	rows, err := pool.Query(ctx,
		artistCols+` WHERE festival_id = $1 AND starts_at > NOW() AND starts_at <= NOW() + INTERVAL '1 hour' ORDER BY starts_at, stage`,
		festivalID,
	)
	if err != nil {
		return nil, fmt.Errorf("next playing: %w", err)
	}
	defer rows.Close()
	return collectArtists(rows)
}

// ByDay returns all sets on the given weekday (in loc) for the given festival.
func ByDay(ctx context.Context, pool *pgxpool.Pool, festivalID int, day time.Weekday, stage string, loc *time.Location) ([]Artist, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if stage != "" {
		rows, err = pool.Query(ctx,
			artistCols+` WHERE festival_id = $1 AND stage ILIKE $2 ORDER BY starts_at`,
			festivalID, stage,
		)
	} else {
		rows, err = pool.Query(ctx,
			artistCols+` WHERE festival_id = $1 ORDER BY stage, starts_at`,
			festivalID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("by day: %w", err)
	}
	defer rows.Close()
	all, err := collectArtists(rows)
	if err != nil {
		return nil, err
	}
	return filterDay(all, day, loc), nil
}

// ByArtist returns sets whose name contains the search string for the given festival.
func ByArtist(ctx context.Context, pool *pgxpool.Pool, festivalID int, name string) ([]Artist, error) {
	rows, err := pool.Query(ctx,
		artistCols+` WHERE festival_id = $1 AND name ILIKE '%' || $2 || '%' ORDER BY starts_at`,
		festivalID, name,
	)
	if err != nil {
		return nil, fmt.Errorf("by artist: %w", err)
	}
	defer rows.Close()
	return collectArtists(rows)
}

// Upsert inserts or updates an artist by (festival_id, name, starts_at).
func Upsert(ctx context.Context, pool *pgxpool.Pool, a Artist) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO artists (festival_id, name, stage, starts_at, ends_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (festival_id, name, starts_at) DO UPDATE
		  SET stage = EXCLUDED.stage, ends_at = EXCLUDED.ends_at`,
		a.FestivalID, a.Name, a.Stage, a.StartsAt, a.EndsAt,
	)
	return err
}

func filterDay(artists []Artist, day time.Weekday, loc *time.Location) []Artist {
	var out []Artist
	for _, a := range artists {
		if a.StartsAt.In(loc).Weekday() == day {
			out = append(out, a)
		}
	}
	return out
}

func collectArtists(rows pgx.Rows) ([]Artist, error) {
	var artists []Artist
	for rows.Next() {
		a, err := scanArtist(rows)
		if err != nil {
			return nil, fmt.Errorf("scan artist: %w", err)
		}
		artists = append(artists, a)
	}
	return artists, rows.Err()
}
