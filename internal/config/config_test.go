package config_test

import (
	"testing"

	"github.com/Discard3700/discord-festival-bot/internal/config"
)

func TestLoad_MissingToken(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "")
	t.Setenv("DATABASE_URL", "postgresql://localhost/test")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DISCORD_TOKEN is missing")
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "tok")
	t.Setenv("DATABASE_URL", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoad_OK(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "mytoken")
	t.Setenv("DATABASE_URL", "postgresql://localhost/test")
	t.Setenv("GUILD_ID", "12345")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DiscordToken != "mytoken" {
		t.Errorf("DiscordToken = %q, want %q", cfg.DiscordToken, "mytoken")
	}
	if cfg.DatabaseURL != "postgresql://localhost/test" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgresql://localhost/test")
	}
	if cfg.GuildID != "12345" {
		t.Errorf("GuildID = %q, want %q", cfg.GuildID, "12345")
	}
}

func TestLoad_GuildIDOptional(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "tok")
	t.Setenv("DATABASE_URL", "postgresql://localhost/test")
	t.Setenv("GUILD_ID", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GuildID != "" {
		t.Errorf("GuildID should be empty, got %q", cfg.GuildID)
	}
}
