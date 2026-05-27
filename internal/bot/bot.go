package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/Discard3700/discord-festival-bot/internal/config"
	"github.com/Discard3700/discord-festival-bot/internal/handlers"
)

type Bot struct {
	session *discordgo.Session
	cfg     *config.Config
	cmds    []*discordgo.ApplicationCommand
}

func New(cfg *config.Config) (*Bot, error) {
	s, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &Bot{session: s, cfg: cfg}, nil
}

func (b *Bot) Start() error {
	b.session.AddHandler(handlers.Router)

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("open session: %w", err)
	}

	registered, err := b.session.ApplicationCommandBulkOverwrite(
		b.session.State.User.ID,
		b.cfg.GuildID,
		handlers.Commands(),
	)
	if err != nil {
		return fmt.Errorf("register commands: %w", err)
	}
	b.cmds = registered
	log.Printf("registered %d slash command(s)", len(b.cmds))
	return nil
}

func (b *Bot) Stop() {
	if _, err := b.session.ApplicationCommandBulkOverwrite(
		b.session.State.User.ID,
		b.cfg.GuildID,
		[]*discordgo.ApplicationCommand{},
	); err != nil {
		log.Printf("deregister commands: %v", err)
	}
	b.session.Close()
}