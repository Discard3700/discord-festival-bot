package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Discard3700/discord-festival-bot/internal/lineup"
	"github.com/Discard3700/discord-festival-bot/internal/reminder"
)

// globalReminderChannelID is set at startup. If empty, reminders post to the
// channel where the /remind add command was run.
var globalReminderChannelID string

func SetReminderChannelID(id string) { globalReminderChannelID = id }

func reminderChannel(i *discordgo.InteractionCreate) string {
	if globalReminderChannelID != "" {
		return globalReminderChannelID
	}
	return i.ChannelID
}

var remindCommand = &discordgo.ApplicationCommand{
	Name:        "remind",
	Description: "Manage set reminders",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Set a reminder before an artist's set",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "artist",
					Description: "Artist name (partial match)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "minutes",
					Description: "How many minutes before the set (default 30)",
					Required:    false,
					MinValue:    ptrFloat(0),
					MaxValue:    300,
				},
				festivalOption,
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List pending reminders",
			Options:     []*discordgo.ApplicationCommandOption{festivalOption},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "cancel",
			Description: "Cancel pending reminder(s) for an artist",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "artist",
					Description: "Artist name (partial match)",
					Required:    true,
				},
				festivalOption,
			},
		},
	},
}

func remindHandler(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool) {
	sub := i.ApplicationCommandData().Options[0]
	switch sub.Name {
	case "add":
		remindAdd(s, i, pool, sub.Options)
	case "list":
		remindList(s, i, pool, sub.Options)
	case "cancel":
		remindCancel(s, i, pool, sub.Options)
	default:
		respond(s, i, "Unknown subcommand.")
	}
}

func remindAdd(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	artistName := optString(opts, "artist")
	minutes := optIntDefault(opts, "minutes", 30)

	fest, err := resolveFestival(pool, opts)
	if err != nil {
		respond(s, i, fmt.Sprintf("❌ %v", err))
		return
	}

	artists, err := lineup.ByArtist(context.Background(), pool, fest.ID, artistName)
	if err != nil {
		log.Printf("remind add: %v", err)
		respond(s, i, "Failed to look up artist.")
		return
	}
	if len(artists) == 0 {
		respond(s, i, fmt.Sprintf("No artist matching %q found at **%s**.", artistName, fest.Name))
		return
	}

	loc := festivalLoc()
	ch := reminderChannel(i)
	ctx := context.Background()

	var created []string
	for _, a := range artists {
		remindAt := a.StartsAt.Add(-time.Duration(minutes) * time.Minute)
		if remindAt.Before(time.Now()) {
			created = append(created, fmt.Sprintf("⚠️ **%s** (%s at %s) — set already started or reminder time passed",
				a.Name, a.Stage, a.StartsAt.In(loc).Format("3:04 PM")))
			continue
		}
		if err := reminder.Create(ctx, pool, a.ID, ch, remindAt); err != nil {
			log.Printf("remind add create: %v", err)
			created = append(created, fmt.Sprintf("❌ **%s** — failed to save reminder", a.Name))
			continue
		}
		created = append(created, fmt.Sprintf("✅ **%s** — %s on %s, reminder at **%s**",
			a.Name, a.Stage,
			a.StartsAt.In(loc).Format("3:04 PM Mon"),
			remindAt.In(loc).Format("3:04 PM"),
		))
	}

	respondEmbed(s, i, &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🔔 Reminder%s Set — %s", pluralS(len(created)), fest.Name),
		Color:       0xFF6B35,
		Description: strings.Join(created, "\n"),
	})
}

func remindList(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	fest, err := resolveFestival(pool, opts)
	if err != nil {
		respond(s, i, fmt.Sprintf("❌ %v", err))
		return
	}

	pending, err := reminder.Pending(context.Background(), pool, fest.ID)
	if err != nil {
		log.Printf("remind list: %v", err)
		respond(s, i, "Failed to fetch reminders.")
		return
	}
	if len(pending) == 0 {
		respond(s, i, fmt.Sprintf("No pending reminders for **%s**.", fest.Name))
		return
	}

	loc := festivalLoc()
	var sb strings.Builder
	for _, r := range pending {
		fmt.Fprintf(&sb, "🔔 **%s** — %s at %s *(reminder at %s)*\n",
			r.ArtistName, r.Stage,
			r.StartsAt.In(loc).Format("3:04 PM Mon"),
			r.RemindAt.In(loc).Format("3:04 PM"),
		)
	}

	respondEmbed(s, i, &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📋 Pending Reminders — %s", fest.Name),
		Color:       0xFF6B35,
		Description: sb.String(),
	})
}

func remindCancel(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	artistName := optString(opts, "artist")

	fest, err := resolveFestival(pool, opts)
	if err != nil {
		respond(s, i, fmt.Sprintf("❌ %v", err))
		return
	}

	artists, err := lineup.ByArtist(context.Background(), pool, fest.ID, artistName)
	if err != nil {
		log.Printf("remind cancel: %v", err)
		respond(s, i, "Failed to look up artist.")
		return
	}
	if len(artists) == 0 {
		respond(s, i, fmt.Sprintf("No artist matching %q found at **%s**.", artistName, fest.Name))
		return
	}

	ctx := context.Background()
	total := int64(0)
	for _, a := range artists {
		n, err := reminder.CancelByArtist(ctx, pool, a.ID)
		if err != nil {
			log.Printf("remind cancel artist %d: %v", a.ID, err)
		}
		total += n
	}

	if total == 0 {
		respond(s, i, fmt.Sprintf("No pending reminders found for %q.", artistName))
		return
	}
	respond(s, i, fmt.Sprintf("✅ Cancelled %d reminder%s for %q.", total, pluralS(int(total)), artistName))
}

func optIntDefault(opts []*discordgo.ApplicationCommandInteractionDataOption, name string, def int) int {
	for _, o := range opts {
		if o.Name == name {
			return int(o.IntValue())
		}
	}
	return def
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func ptrFloat(v float64) *float64 { return &v }
