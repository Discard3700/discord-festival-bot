package handlers_test

import (
	"testing"

	"github.com/Discard3700/discord-festival-bot/internal/handlers"
)

func TestLineupCommandRegistered(t *testing.T) {
	for _, c := range handlers.Commands() {
		if c.Name == "lineup" {
			return
		}
	}
	t.Fatal("lineup command not found in Commands()")
}

func TestLineupCommandSubcommands(t *testing.T) {
	var cmd *struct{ Options interface{} }
	for _, c := range handlers.Commands() {
		if c.Name == "lineup" {
			want := []string{"now", "next", "day", "artist"}
			got := map[string]bool{}
			for _, o := range c.Options {
				got[o.Name] = true
			}
			for _, name := range want {
				if !got[name] {
					t.Errorf("lineup command missing subcommand %q", name)
				}
			}
			return
		}
	}
	_ = cmd
	t.Fatal("lineup command not found")
}

func TestLineupDaySubcommandOptions(t *testing.T) {
	for _, c := range handlers.Commands() {
		if c.Name != "lineup" {
			continue
		}
		for _, sub := range c.Options {
			if sub.Name != "day" {
				continue
			}
			opts := map[string]bool{}
			for _, o := range sub.Options {
				opts[o.Name] = true
			}
			if !opts["day"] {
				t.Error("lineup day subcommand missing 'day' option")
			}
			if !opts["stage"] {
				t.Error("lineup day subcommand missing 'stage' option")
			}
			return
		}
	}
	t.Fatal("lineup day subcommand not found")
}
