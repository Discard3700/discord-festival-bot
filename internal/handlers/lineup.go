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
)

var festivalOption = &discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionString,
	Name:        "festival",
	Description: "Festival slug (defaults to active/upcoming festival)",
	Required:    false,
}

var lineupCommand = &discordgo.ApplicationCommand{
	Name:        "lineup",
	Description: "Browse the festival lineup",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "now",
			Description: "What's playing right now",
			Options:     []*discordgo.ApplicationCommandOption{festivalOption},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "next",
			Description: "Sets starting in the next hour",
			Options:     []*discordgo.ApplicationCommandOption{festivalOption},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "day",
			Description: "All sets on a given festival day",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "day",
					Description: "Day name (e.g. Friday, Saturday)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "stage",
					Description: "Filter by stage name",
					Required:    false,
				},
				festivalOption,
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "artist",
			Description: "Look up a specific artist's set time",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Artist name (partial match)",
					Required:    true,
				},
				festivalOption,
			},
		},
	},
}

func lineupHandler(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool) {
	sub := i.ApplicationCommandData().Options[0]
	switch sub.Name {
	case "now":
		lineupNow(s, i, pool, sub.Options)
	case "next":
		lineupNext(s, i, pool, sub.Options)
	case "day":
		lineupDay(s, i, pool, sub.Options)
	case "artist":
		lineupArtist(s, i, pool, sub.Options)
	default:
		respond(s, i, "Unknown subcommand.")
	}
}

// resolveFestival looks up by slug if given, otherwise returns the default.
func resolveFestival(pool *pgxpool.Pool, opts []*discordgo.ApplicationCommandInteractionDataOption) (*lineup.Festival, error) {
	slug := optString(opts, "festival")
	if slug != "" {
		return lineup.FestivalBySlug(context.Background(), pool, slug)
	}
	return lineup.DefaultFestival(context.Background(), pool)
}

func lineupNow(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	fest, err := resolveFestival(pool, opts)
	if err != nil {
		respond(s, i, fmt.Sprintf("❌ %v", err))
		return
	}
	artists, err := lineup.NowPlaying(context.Background(), pool, fest.ID)
	if err != nil {
		log.Printf("lineup now: %v", err)
		respond(s, i, "Failed to fetch lineup.")
		return
	}
	if len(artists) == 0 {
		respond(s, i, fmt.Sprintf("Nothing is playing right now at **%s**.", fest.Name))
		return
	}
	loc := festivalLoc(pool)
	respondEmbed(s, i, &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🎵 Now Playing — %s", fest.Name),
		Color: 0x1DB954,
		Description: formatList(artists, func(a lineup.Artist) string {
			return fmt.Sprintf("**%s** — %s *(until %s)*", a.Name, a.Stage, fmtTime(a.EndsAt, loc))
		}),
	})
}

func lineupNext(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	fest, err := resolveFestival(pool, opts)
	if err != nil {
		respond(s, i, fmt.Sprintf("❌ %v", err))
		return
	}
	artists, err := lineup.NextPlaying(context.Background(), pool, fest.ID)
	if err != nil {
		log.Printf("lineup next: %v", err)
		respond(s, i, "Failed to fetch lineup.")
		return
	}
	if len(artists) == 0 {
		respond(s, i, fmt.Sprintf("Nothing starting in the next hour at **%s**.", fest.Name))
		return
	}
	loc := festivalLoc(pool)
	respondEmbed(s, i, &discordgo.MessageEmbed{
		Title: fmt.Sprintf("⏭️ Up Next — %s", fest.Name),
		Color: 0xF59E0B,
		Description: formatList(artists, func(a lineup.Artist) string {
			return fmt.Sprintf("**%s** — %s *(at %s)*", a.Name, a.Stage, fmtTime(a.StartsAt, loc))
		}),
	})
}

func lineupDay(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	dayStr := optString(opts, "day")
	stage := optString(opts, "stage")

	day, ok := lineup.ParseWeekday(dayStr)
	if !ok {
		respond(s, i, fmt.Sprintf("Unknown day %q. Try Friday, Saturday, etc.", dayStr))
		return
	}

	fest, err := resolveFestival(pool, opts)
	if err != nil {
		respond(s, i, fmt.Sprintf("❌ %v", err))
		return
	}

	loc := festivalLoc(pool)
	artists, err := lineup.ByDay(context.Background(), pool, fest.ID, day, stage, loc)
	if err != nil {
		log.Printf("lineup day: %v", err)
		respond(s, i, "Failed to fetch lineup.")
		return
	}

	dayTitle := capitalize(dayStr)
	if len(artists) == 0 {
		msg := fmt.Sprintf("No sets found for **%s** at %s.", dayTitle, fest.Name)
		if stage != "" {
			msg = fmt.Sprintf("No sets found for **%s** on %s at %s.", dayTitle, stage, fest.Name)
		}
		respond(s, i, msg)
		return
	}

	title := fmt.Sprintf("📅 %s — %s", dayTitle, fest.Name)
	if stage != "" {
		title = fmt.Sprintf("📅 %s · %s — %s", dayTitle, stage, fest.Name)
	}
	respondEmbed(s, i, &discordgo.MessageEmbed{
		Title:  title,
		Color:  0x6366F1,
		Fields: groupByStage(artists, loc),
	})
}

func lineupArtist(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	name := optString(opts, "name")

	fest, err := resolveFestival(pool, opts)
	if err != nil {
		respond(s, i, fmt.Sprintf("❌ %v", err))
		return
	}

	artists, err := lineup.ByArtist(context.Background(), pool, fest.ID, name)
	if err != nil {
		log.Printf("lineup artist: %v", err)
		respond(s, i, "Failed to fetch lineup.")
		return
	}
	if len(artists) == 0 {
		respond(s, i, fmt.Sprintf("No artists matching %q at **%s**.", name, fest.Name))
		return
	}

	loc := festivalLoc(pool)
	respondEmbed(s, i, &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🎤 Artist Search — %s", fest.Name),
		Color: 0xEC4899,
		Description: formatList(artists, func(a lineup.Artist) string {
			return fmt.Sprintf("**%s** — %s\n%s – %s (%s)",
				a.Name, a.Stage,
				fmtTime(a.StartsAt, loc), fmtTime(a.EndsAt, loc),
				a.StartsAt.In(loc).Format("Monday"),
			)
		}),
	})
}

func groupByStage(artists []lineup.Artist, loc *time.Location) []*discordgo.MessageEmbedField {
	order := []string{}
	groups := map[string][]lineup.Artist{}
	for _, a := range artists {
		if _, seen := groups[a.Stage]; !seen {
			order = append(order, a.Stage)
		}
		groups[a.Stage] = append(groups[a.Stage], a)
	}

	fields := make([]*discordgo.MessageEmbedField, 0, len(order))
	for _, stage := range order {
		var sb strings.Builder
		for _, a := range groups[stage] {
			fmt.Fprintf(&sb, "**%s** %s–%s\n", a.Name, fmtTime(a.StartsAt, loc), fmtTime(a.EndsAt, loc))
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  stage,
			Value: sb.String(),
		})
	}
	return fields
}

func formatList(artists []lineup.Artist, line func(lineup.Artist) string) string {
	var sb strings.Builder
	for _, a := range artists {
		sb.WriteString(line(a))
		sb.WriteByte('\n')
	}
	return sb.String()
}

func fmtTime(t time.Time, loc *time.Location) string {
	return t.In(loc).Format("3:04 PM")
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

var globalLoc *time.Location

func SetFestivalLoc(loc *time.Location) { globalLoc = loc }

func festivalLoc(_ *pgxpool.Pool) *time.Location {
	if globalLoc != nil {
		return globalLoc
	}
	return time.UTC
}

func optString(opts []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, o := range opts {
		if o.Name == name {
			return o.StringValue()
		}
	}
	return ""
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	}); err != nil {
		log.Printf("respond: %v", err)
	}
}

func respondEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
	}); err != nil {
		log.Printf("respondEmbed: %v", err)
	}
}
