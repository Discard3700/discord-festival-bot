package reminder_test

import (
	"testing"
	"time"

	"github.com/Discard3700/discord-festival-bot/internal/reminder"
)

func TestPollerCreation(t *testing.T) {
	p := reminder.NewPoller(nil, nil, time.UTC)
	if p == nil {
		t.Fatal("NewPoller returned nil")
	}
}
