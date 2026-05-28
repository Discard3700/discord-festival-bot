package handlers_test

import (
	"testing"

	"github.com/Discard3700/discord-festival-bot/internal/handlers"
)

func TestCommands_NotEmpty(t *testing.T) {
	if len(handlers.Commands()) == 0 {
		t.Fatal("Commands() returned empty slice — at least /ping must be registered")
	}
}

func TestCommands_NamesUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, c := range handlers.Commands() {
		if seen[c.Name] {
			t.Errorf("duplicate command name: %q", c.Name)
		}
		seen[c.Name] = true
	}
}

func TestCommands_ContainsPing(t *testing.T) {
	for _, c := range handlers.Commands() {
		if c.Name == "ping" {
			return
		}
	}
	t.Fatal("ping command not found in Commands()")
}

func TestRouter_ReturnsFunc(t *testing.T) {
	// nil pool is intentional — Router must not dereference it at construction time
	if handlers.Router(nil) == nil {
		t.Fatal("Router(nil) returned nil handler")
	}
}
