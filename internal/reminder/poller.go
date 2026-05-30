package reminder

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Poller checks for due reminders every 30 s and posts them to Discord.
type Poller struct {
	session *discordgo.Session
	pool    *pgxpool.Pool
	loc     *time.Location
	done    chan struct{}
}

func NewPoller(session *discordgo.Session, pool *pgxpool.Pool, loc *time.Location) *Poller {
	return &Poller{session: session, pool: pool, loc: loc, done: make(chan struct{})}
}

func (p *Poller) Start() {
	go p.run()
}

func (p *Poller) Stop() {
	close(p.done)
}

func (p *Poller) run() {
	p.tick() // fire immediately to catch anything missed during downtime
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.tick()
		case <-p.done:
			return
		}
	}
}

func (p *Poller) tick() {
	ctx := context.Background()
	due, err := Due(ctx, p.pool)
	if err != nil {
		log.Printf("reminder poller: %v", err)
		return
	}
	for _, r := range due {
		if err := p.post(r); err != nil {
			log.Printf("reminder post id=%d: %v", r.ID, err)
			continue
		}
		if err := MarkSent(ctx, p.pool, r.ID); err != nil {
			log.Printf("reminder mark sent id=%d: %v", r.ID, err)
		}
	}
}

func (p *Poller) post(r WithArtist) error {
	minsUntil := int(time.Until(r.StartsAt).Minutes())

	var title, desc string
	switch {
	case minsUntil <= 0:
		title = "🎵 On Stage Now!"
		desc = fmt.Sprintf("**%s** is playing on **%s** until **%s**",
			r.ArtistName, r.Stage, r.EndsAt.In(p.loc).Format("3:04 PM"))
	case minsUntil == 1:
		title = "🔔 Starting in 1 minute"
		desc = fmt.Sprintf("**%s** — %s at **%s**",
			r.ArtistName, r.Stage, r.StartsAt.In(p.loc).Format("3:04 PM"))
	default:
		title = fmt.Sprintf("🔔 Starting in %d minutes", minsUntil)
		desc = fmt.Sprintf("**%s** — %s at **%s**",
			r.ArtistName, r.Stage, r.StartsAt.In(p.loc).Format("3:04 PM"))
	}

	_, err := p.session.ChannelMessageSendEmbed(r.ChannelID, &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Color:       0xFF6B35,
		Footer:      &discordgo.MessageEmbedFooter{Text: r.FestivalName},
	})
	return err
}
