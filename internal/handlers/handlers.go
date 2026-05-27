package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type handler func(s *discordgo.Session, i *discordgo.InteractionCreate)

var registry = map[string]handler{
	"ping": pingHandler,
}

var commandDefs = []*discordgo.ApplicationCommand{
	pingCommand,
}

// Commands returns the slice of command definitions to register with Discord.
func Commands() []*discordgo.ApplicationCommand {
	return commandDefs
}

// Router dispatches incoming interactions to the registered handler.
func Router(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	name := i.ApplicationCommandData().Name
	h, ok := registry[name]
	if !ok {
		log.Printf("unknown command: %s", name)
		return
	}
	h(s, i)
}