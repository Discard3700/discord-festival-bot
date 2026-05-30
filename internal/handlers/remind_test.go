package handlers_test

import (
	"testing"

	"github.com/Discard3700/discord-festival-bot/internal/handlers"
)

func TestRemindCommandRegistered(t *testing.T) {
	for _, c := range handlers.Commands() {
		if c.Name == "remind" {
			return
		}
	}
	t.Fatal("remind command not found in Commands()")
}

func TestRemindSubcommands(t *testing.T) {
	for _, c := range handlers.Commands() {
		if c.Name != "remind" {
			continue
		}
		want := []string{"add", "list", "cancel"}
		got := map[string]bool{}
		for _, o := range c.Options {
			got[o.Name] = true
		}
		for _, name := range want {
			if !got[name] {
				t.Errorf("remind command missing subcommand %q", name)
			}
		}
		return
	}
	t.Fatal("remind command not found")
}

func TestRemindAddOptions(t *testing.T) {
	for _, c := range handlers.Commands() {
		if c.Name != "remind" {
			continue
		}
		for _, sub := range c.Options {
			if sub.Name != "add" {
				continue
			}
			opts := map[string]bool{}
			for _, o := range sub.Options {
				opts[o.Name] = true
			}
			if !opts["artist"] {
				t.Error("remind add missing 'artist' option")
			}
			if !opts["minutes"] {
				t.Error("remind add missing 'minutes' option")
			}
			if !opts["festival"] {
				t.Error("remind add missing 'festival' option")
			}
			return
		}
	}
	t.Fatal("remind add subcommand not found")
}
