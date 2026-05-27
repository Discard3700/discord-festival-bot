package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/Discard3700/discord-festival-bot/internal/bot"
	"github.com/Discard3700/discord-festival-bot/internal/config"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("bot: %v", err)
	}

	if err := b.Start(); err != nil {
		log.Fatalf("bot start: %v", err)
	}
	defer b.Stop()

	log.Println("bot is running — press Ctrl+C to exit")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}