package config

import (
	"fmt"
	"os"
)

type Config struct {
	DiscordToken string
	GuildID      string
	DatabaseURL  string
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

	return &Config{
		DiscordToken: token,
		GuildID:      os.Getenv("GUILD_ID"),
		DatabaseURL:  dbURL,
	}, nil
}
