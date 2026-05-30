package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

var countdownCommand = &discordgo.ApplicationCommand{
	Name:        "countdown",
	Description: "Time until the festival starts (or how it's going)",
	Options:     []*discordgo.ApplicationCommandOption{festivalOption},
}

func countdownHandler(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool) {
	fest, err := resolveFestival(pool, i.ApplicationCommandData().Options)
	if err != nil {
		respond(s, i, fmt.Sprintf("❌ %v", err))
		return
	}

	now := time.Now()
	loc := festivalLoc()

	var title, desc string
	var color int

	switch {
	case now.Before(fest.StartsAt):
		until := fest.StartsAt.Sub(now)
		title = fmt.Sprintf("⏳ %s", fest.Name)
		desc = fmt.Sprintf("Starts in **%s**\n_%s_",
			FmtDuration(until),
			fest.StartsAt.In(loc).Format("Monday, Jan 2 at 3:04 PM"),
		)
		color = 0x6366F1

	case now.After(fest.EndsAt):
		since := now.Sub(fest.EndsAt)
		title = fmt.Sprintf("🎪 %s", fest.Name)
		desc = fmt.Sprintf("Ended **%s** ago\n_%s_",
			FmtDuration(since),
			fest.EndsAt.In(loc).Format("Monday, Jan 2"),
		)
		color = 0x6B7280

	default:
		elapsed := now.Sub(fest.StartsAt)
		remaining := fest.EndsAt.Sub(now)
		total := fest.EndsAt.Sub(fest.StartsAt)
		pct := int(float64(elapsed) / float64(total) * 100)

		title = fmt.Sprintf("🎉 %s is LIVE!", fest.Name)
		desc = fmt.Sprintf("%s **%d%%**\nStarted **%s** ago · **%s** to go",
			progressBar(elapsed, total), pct,
			FmtDuration(elapsed),
			FmtDuration(remaining),
		)
		color = 0x1DB954
	}

	respondEmbed(s, i, &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Color:       color,
	})
}

func FmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d day%s", days, pluralS(days)))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hr%s", hours, pluralS(hours)))
	}
	if mins > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d min%s", mins, pluralS(mins)))
	}
	return strings.Join(parts, " ")
}

func progressBar(elapsed, total time.Duration) string {
	const width = 10
	filled := int(float64(elapsed) / float64(total) * width)
	if filled > width {
		filled = width
	}
	return strings.Repeat("▓", filled) + strings.Repeat("░", width-filled)
}
