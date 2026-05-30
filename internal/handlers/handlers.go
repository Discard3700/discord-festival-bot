package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type handler func(s *discordgo.Session, i *discordgo.InteractionCreate, pool *pgxpool.Pool)

var registry = map[string]handler{
	"ping":   pingHandler,
	"lineup": lineupHandler,
	"remind": remindHandler,
}

var commandDefs = []*discordgo.ApplicationCommand{
	pingCommand,
	lineupCommand,
	remindCommand,
}

func Commands() []*discordgo.ApplicationCommand {
	return commandDefs
}

// Router returns a discordgo-compatible handler with the pool closed over.
func Router(pool *pgxpool.Pool) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		name := i.ApplicationCommandData().Name
		h, ok := registry[name]
		if !ok {
			log.Printf("unknown command: %s", name)
			return
		}
		h(s, i, pool)
	}
}
