package main

import (
	"log"
	"os"
	"os/signal"

	"example.com/discord-ai/pkg/ai"
	"github.com/alecthomas/kong"
	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

var CLI struct {
	Run Run `cmd:"" help:"Run the discord bot"`
}

type Run struct {
	Token       string `help:"API key for discord" env:"DISCORD_TOKEN"`
	OpenAIToken string `help:"API key for openai" env:"OPEN_API_KEY"`
	GuildID     string `help:"guild id of the discord server" env:"GUILD_ID"`
	ChannelID   string `help:"channel id that the bot should join" env:"CHANNEL_ID"`
}

func main() {
	ctx := kong.Parse(&CLI)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func (r *Run) Run() error {
	openAIClient := openai.NewClient(r.OpenAIToken)
	discord, err := discordgo.New("Bot " + r.Token)
	if err != nil {
		return err
	}

	if err := discord.Open(); err != nil {
		return err
	}
	defer discord.Close()

	aiWrapper := ai.New(discord, openAIClient)
	removeCmds := registerCommands(discord, aiWrapper, r.GuildID)
	defer removeCmds()

	go aiWrapper.Work()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
	return nil
}
