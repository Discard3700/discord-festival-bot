package handlers_test

import (
	"testing"
	"time"

	"github.com/Discard3700/discord-festival-bot/internal/handlers"
)

func TestCountdownCommandRegistered(t *testing.T) {
	for _, c := range handlers.Commands() {
		if c.Name == "countdown" {
			return
		}
	}
	t.Fatal("countdown command not found in Commands()")
}

func TestFmtDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{0, "0 mins"},
		{time.Minute, "1 min"},
		{2 * time.Minute, "2 mins"},
		{60 * time.Minute, "1 hr"},
		{90 * time.Minute, "1 hr 30 mins"},
		{24 * time.Hour, "1 day"},
		{25*time.Hour + 30*time.Minute, "1 day 1 hr 30 mins"},
		{48 * time.Hour, "2 days"},
	}
	for _, c := range cases {
		got := handlers.FmtDuration(c.d)
		if got != c.want {
			t.Errorf("FmtDuration(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}
