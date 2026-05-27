package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var pingCommand = &discordgo.ApplicationCommand{
	Name:        "ping",
	Description: "Check that the bot is alive",
}

func pingHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🏓 Pong!",
		},
	})
	if err != nil {
		log.Printf("ping respond: %v", err)
	}
}