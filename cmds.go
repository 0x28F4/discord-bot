package main

import (
	"fmt"
	"net/http"

	"github.com/0x28F4/discord-bot/pkg/ai"
	"github.com/0x28F4/discord-bot/pkg/elevenlabs"
	"github.com/0x28F4/discord-bot/pkg/request"
	"github.com/bwmarrin/discordgo"
)

func makeOptionMap(i *discordgo.InteractionCreate) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	return optionMap
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, text string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	})
}

func registerCommands(s *discordgo.Session, aiWrapper *ai.AI, guildID string, elevenlabsAPIKey string) (removeCmds func()) {
	voices := makeVoiceChoices()
	if elevenlabsAPIKey != "" {
		voices = makeElevenlabsVoices(elevenlabsAPIKey)
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "say",
			Description: "make the bot say something in a voice channel",
			Options: []*discordgo.ApplicationCommandOption{
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
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "whats your question senpai?",
					Required:    true,
				},
			},
		},
		{
			Name:        "join",
			Description: "make the bot join a voice channel",
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
			},
		},
		{
			Name:        "leave",
			Description: "make the bot leave the current voice channel",
		},
		{
			Name:        "voice",
			Description: "change voice to something else",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Choices:     voices,
					Name:        "voice",
					Description: "pick the voice of the bot",
					Required:    true,
				},
			},
		},
	}

	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"say": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !aiWrapper.Discord.Connected() {
				respond(s, i, "the bot isn't connected to any voice channel, please use the /join command first")
				return
			}
			respond(s, i, fmt.Sprintf("processing command, queue length: %d", aiWrapper.QueueLength()))

			optionMap := makeOptionMap(i)
			text, hasText := optionMap["text"].Value.(string)
			if !hasText {
				fmt.Printf("no text was given")
			}

			aiWrapper.Queue(&ai.SayCmd{
				Prompt:  text,
				GuildId: guildID,
			})
		},
		"ask": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !aiWrapper.Discord.Connected() {
				respond(s, i, "the bot isn't connected to any voice channel, please use the /join command first")
				return
			}
			respond(s, i, fmt.Sprintf("processing command, queue length: %d", aiWrapper.QueueLength()))

			optionMap := makeOptionMap(i)
			text, hasText := optionMap["text"].Value.(string)
			if !hasText {
				fmt.Printf("no text was given")
			}

			aiWrapper.Queue(&ai.AskCmd{
				Prompt:  text,
				GuildId: guildID,
			})
		},
		"join": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			respond(s, i, "I'm joining the voice channel senpai")
			optionMap := makeOptionMap(i)
			channelId, hasChannelId := optionMap["channel"].Value.(string)
			if !hasChannelId {
				fmt.Printf("no channel id given")
				return
			}
			if err := aiWrapper.Discord.JoinChannel(channelId); err != nil {
				fmt.Println(err)
			}
		},
		"leave": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			respond(s, i, "I'm leaving the voice channel senpai")
			if err := aiWrapper.Discord.LeaveChannel(); err != nil {
				fmt.Println(err)
			}
		},
		"voice": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			respond(s, i, "I'm changing my voice, do you prefer male voice? That's kinda sus")
			optionMap := makeOptionMap(i)
			if voice, hasVoice := optionMap["voice"].Value.(string); hasVoice {
				aiWrapper.TTS.ChangeVoice(voice)
			}
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
			fmt.Printf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	return func() {
		fmt.Println("performing cleanup tasks")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, guildID, v.ID)
			if err != nil {
				fmt.Printf("Cannot delete '%v' command: %v\n", v.Name, err)
			}
		}

		if aiWrapper.Discord.Connected() {
			if err := aiWrapper.Discord.LeaveChannel(); err != nil {
				fmt.Printf("cannot leave current voice channel: %v\n", err)
			}
		}
	}
}

func makeVoiceChoices() []*discordgo.ApplicationCommandOptionChoice {
	var voices = []string{
		"p225", "p226", "p227", "p228", "p229", "p230", "p231", "p232", "p233", "p234", "p236", "p237", "p238", "p239", "p240", "p241", "p243", "p244", "p245", "p246", "p247", "p248", "p249", "p250", "p251", "p252", "p253", "p254", "p255", "p256", "p257", "p258", "p259", "p260", "p261", "p262", "p263", "p264", "p265", "p266", "p267", "p268", "p269", "p270", "p271", "p272", "p273", "p274", "p275", "p276", "p277", "p278", "p279", "p280", "p281", "p282", "p283", "p284", "p285", "p286", "p287", "p288", "p292", "p293", "p294", "p295", "p297", "p298", "p299", "p300", "p301", "p302", "p303", "p304", "p305", "p306", "p307", "p308", "p310", "p311", "p312", "p313", "p314", "p316", "p317", "p318", "p323", "p326", "p329", "p330", "p333", "p334", "p335", "p336", "p339", "p340", "p341", "p343", "p345", "p347", "p351", "p360", "p361", "p362", "p363", "p364", "p374", "p376",
	}

	options := make([]*discordgo.ApplicationCommandOptionChoice, 25)
	for i := range options {
		v := voices[i]
		options[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  v,
			Value: v,
		}
	}

	return options
}

func makeElevenlabsVoices(apiKey string) []*discordgo.ApplicationCommandOptionChoice {
	headers := map[string]string{
		"xi-api-key": apiKey,
	}
	result, err := request.Request[elevenlabs.VoicesResponse](http.MethodGet, "https://api.elevenlabs.io/v1/voices", headers, nil)
	if err != nil {
		panic(fmt.Sprintf("couldn't retrieve voices from elevenlabs: %v", err))
	}

	options := make([]*discordgo.ApplicationCommandOptionChoice, len(result.Voices))
	for i := range options {
		voice := result.Voices[i]
		options[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  voice.Name,
			Value: voice.VoiceID,
		}
	}

	return options
}
