package main

import (
	"fmt"
	"log"

	"example.com/discord-ai/pkg/ai"
	"github.com/bwmarrin/discordgo"
)

func registerCommands(s *discordgo.Session, aiWrapper *ai.AI, guildID string) (removeCmds func()) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "say",
			Description: "make the bot say something in a voice channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "Channel",
					ChannelTypes: []discordgo.ChannelType{
						discordgo.ChannelTypeGuildVoice,
					},
					Required: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "what do you have to say buddy?",
					Required:    true,
				},
			},
		},
		{
			Name:        "ask",
			Description: "ask the bot something",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "Channel",
					ChannelTypes: []discordgo.ChannelType{
						discordgo.ChannelTypeGuildVoice,
					},
					Required: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "whats your question senpai?",
					Required:    true,
				},
			},
		},
	}

	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"say": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "processing",
				},
			})
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			channelId, hasChannelId := optionMap["channel"].Value.(string)
			text, hasText := optionMap["text"].Value.(string)
			if !hasChannelId || !hasText {
				fmt.Printf("either no channel or no text was given")
			}

			aiWrapper.Queue(&ai.SayCmd{
				Prompt:    text,
				GuildId:   guildID,
				ChannelId: channelId,
			})
		},
		"ask": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "processing",
				},
			})
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			channelId, hasChannelId := optionMap["channel"].Value.(string)
			text, hasText := optionMap["text"].Value.(string)
			if !hasChannelId || !hasText {
				fmt.Printf("either no channel or no text was given")
			}

			aiWrapper.Queue(&ai.AskCmd{
				Prompt:    text,
				GuildId:   guildID,
				ChannelId: channelId,
			})
		},
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	return func() {
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, guildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}
}
