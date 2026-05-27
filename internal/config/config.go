package config

import (
	"fmt"
	"os"
)

type Config struct {
	DiscordToken string
	GuildID      string // empty = global commands (slower to propagate)
}

func Load() (*Config, error) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is required")
	}

	return &Config{
		DiscordToken: token,
		GuildID:      os.Getenv("GUILD_ID"),
	}, nil
}