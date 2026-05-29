package reminder

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithArtist is a reminder joined with its artist and festival.
type WithArtist struct {
	ID           int
	ChannelID    string
	RemindAt     time.Time
	ArtistID     int
	ArtistName   string
	Stage        string
	StartsAt     time.Time
	EndsAt       time.Time
	FestivalName string
}

const joinCols = `
	SELECT r.id, r.channel_id, r.remind_at,
	       a.id, a.name, a.stage, a.starts_at, a.ends_at,
	       f.name
	FROM reminders r
	JOIN artists  a ON a.id  = r.artist_id
	JOIN festivals f ON f.id = a.festival_id`

func scanRow(row pgx.Row) (WithArtist, error) {
	var r WithArtist
	err := row.Scan(&r.ID, &r.ChannelID, &r.RemindAt,
		&r.ArtistID, &r.ArtistName, &r.Stage, &r.StartsAt, &r.EndsAt,
		&r.FestivalName)
	return r, err
}

func collectRows(rows pgx.Rows) ([]WithArtist, error) {
	var out []WithArtist
	for rows.Next() {
		var r WithArtist
		if err := rows.Scan(&r.ID, &r.ChannelID, &r.RemindAt,
			&r.ArtistID, &r.ArtistName, &r.Stage, &r.StartsAt, &r.EndsAt,
			&r.FestivalName); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Create inserts a reminder for an artist.
func Create(ctx context.Context, pool *pgxpool.Pool, artistID int, channelID string, remindAt time.Time) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO reminders (artist_id, channel_id, remind_at) VALUES ($1, $2, $3)`,
		artistID, channelID, remindAt,
	)
	return err
}

// Due returns all unsent reminders whose remind_at has passed.
func Due(ctx context.Context, pool *pgxpool.Pool) ([]WithArtist, error) {
	rows, err := pool.Query(ctx,
		joinCols+` WHERE NOT r.sent AND r.remind_at <= NOW() ORDER BY r.remind_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("due reminders: %w", err)
	}
	defer rows.Close()
	return collectRows(rows)
}

// MarkSent marks a reminder as delivered.
func MarkSent(ctx context.Context, pool *pgxpool.Pool, id int) error {
	_, err := pool.Exec(ctx, `UPDATE reminders SET sent = TRUE WHERE id = $1`, id)
	return err
}

// Pending returns unsent reminders for a festival, ordered by remind_at.
func Pending(ctx context.Context, pool *pgxpool.Pool, festivalID int) ([]WithArtist, error) {
	rows, err := pool.Query(ctx,
		joinCols+` WHERE NOT r.sent AND f.id = $1 ORDER BY r.remind_at`,
		festivalID,
	)
	if err != nil {
		return nil, fmt.Errorf("pending reminders: %w", err)
	}
	defer rows.Close()
	return collectRows(rows)
}

// CancelByArtist removes all unsent reminders for a given artist.
func CancelByArtist(ctx context.Context, pool *pgxpool.Pool, artistID int) (int64, error) {
	tag, err := pool.Exec(ctx,
		`DELETE FROM reminders WHERE artist_id = $1 AND NOT sent`, artistID,
	)
	return tag.RowsAffected(), err
}

// ForArtist returns unsent reminders for a specific artist.
func ForArtist(ctx context.Context, pool *pgxpool.Pool, artistID int) ([]WithArtist, error) {
	rows, err := pool.Query(ctx,
		joinCols+` WHERE NOT r.sent AND r.artist_id = $1 ORDER BY r.remind_at`,
		artistID,
	)
	if err != nil {
		return nil, fmt.Errorf("reminders for artist: %w", err)
	}
	defer rows.Close()
	return collectRows(rows)
}

// unused but satisfies import of scanRow for potential future single-row use
var _ = scanRow
