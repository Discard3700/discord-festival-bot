package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DiscordToken string
	GuildID      string
	DatabaseURL  string
	FestivalTZ   *time.Location
}

func Load() (*Config, error) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is required")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	tzName := os.Getenv("FESTIVAL_TZ")
	if tzName == "" {
		tzName = "UTC"
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("invalid FESTIVAL_TZ %q: %w", tzName, err)
	}

	return &Config{
		DiscordToken: token,
		GuildID:      os.Getenv("GUILD_ID"),
		DatabaseURL:  dbURL,
		FestivalTZ:   loc,
	}, nil
}
