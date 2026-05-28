package lineup_test

import (
	"testing"
	"time"

	"github.com/Discard3700/discord-festival-bot/internal/lineup"
)

func TestParseWeekday(t *testing.T) {
	cases := []struct {
		input string
		want  time.Weekday
		ok    bool
	}{
		{"Friday", time.Friday, true},
		{"friday", time.Friday, true},
		{"fri", time.Friday, true},
		{"FRI", time.Friday, true},
		{"Saturday", time.Saturday, true},
		{"sat", time.Saturday, true},
		{"Sunday", time.Sunday, true},
		{"sun", time.Sunday, true},
		{"Monday", time.Monday, true},
		{"mon", time.Monday, true},
		{"Tuesday", time.Tuesday, true},
		{"tue", time.Tuesday, true},
		{"Wednesday", time.Wednesday, true},
		{"wed", time.Wednesday, true},
		{"Thursday", time.Thursday, true},
		{"thu", time.Thursday, true},
		{"", time.Sunday, false},
		{"funday", time.Sunday, false},
		{"  Friday  ", time.Friday, true},
	}

	for _, c := range cases {
		got, ok := lineup.ParseWeekday(c.input)
		if ok != c.ok {
			t.Errorf("ParseWeekday(%q) ok=%v, want %v", c.input, ok, c.ok)
			continue
		}
		if ok && got != c.want {
			t.Errorf("ParseWeekday(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestFilterDayAlignment(t *testing.T) {
	loc, _ := time.LoadLocation("America/Chicago")
	friday := time.Date(2025, 7, 11, 20, 0, 0, 0, loc)
	if friday.Weekday() != time.Friday {
		t.Errorf("test date is not a Friday: %v", friday.Weekday())
	}
}
