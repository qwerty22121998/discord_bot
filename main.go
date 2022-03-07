package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/qwerty22121998/discord_bot/handlers"
	"github.com/qwerty22121998/discord_bot/parser"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}

func main() {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	discordToken := os.Getenv("DISCORD_TOKEN")
	d, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		panic(err)
	}
	if err := parser.InitParser(apiKey); err != nil {
		panic(err)
	}
	h := handlers.NewMusicHandler()
	go h.Start()

	d.AddHandler(h.Handle)

	d.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentGuildVoiceStates

	err = d.Open()
	if err != nil {
		panic(err)
	}

	zap.S().Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	d.Close()
}
