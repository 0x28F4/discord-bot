package ai

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/0x28F4/discord-bot/pkg/prompts"
	"github.com/0x28F4/discord-bot/pkg/tts"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

type Cmd interface {
	Do(ai *AI) error
}

type AI struct {
	ttsAddress   string
	openAIClient *openai.Client
	queue        []Cmd
	TTS          tts.TTS
	Discord      *discordChannelSession
}

func (a *AI) Queue(cmd Cmd) {
	a.queue = append(a.queue, cmd)
}

func (a *AI) QueueLength() int {
	return len(a.queue)
}

func (a *AI) doNext() (done bool, err error) {
	if len(a.queue) == 0 {
		return true, nil
	}

	current := a.queue[0]
	a.queue = a.queue[1:]
	return false, current.Do(a)
}

func (a *AI) Work() {
	fmt.Println("ai starting work")
	for {
		done, err := a.doNext()
		if err != nil {
			fmt.Printf("couldn't do command, got error: %v\n", err)
		}

		if done {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

type SayCmd struct {
	Prompt    string
	GuildId   string
	ChannelId string
}

func (s *SayCmd) Do(ai *AI) error {
	if err := ai.say(s.Prompt); err != nil {
		return fmt.Errorf("couldn't say \"%s\",err:%v\n", s.Prompt, err)
	}
	return nil
}

type AskCmd struct {
	Prompt    string
	GuildId   string
	ChannelId string
	Mode      string
}

func (a *AskCmd) Do(ai *AI) error {
	output, err := ai.think(a.Prompt, a.Mode)
	if err != nil {
		return fmt.Errorf("couldn't come up with an answer for %s, reason: %v\n", a.Prompt, err)
	} else if err := ai.say(output); err != nil {
		return fmt.Errorf("couldn't say \"%s\"\n", output)
	}

	return nil
}

func (ai *AI) think(prompt, mode string) (out string, err error) {
	p, hasMode := prompts.Prompts[mode]
	if !hasMode {
		return "", fmt.Errorf("couldn't find prompt for \"%s\"", mode)
	}

	resp, err := ai.openAIClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: p.Prompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("think error: %w", err)
	}

	return p.Format(resp.Choices[0].Message.Content), nil
}

func (ai *AI) say(text string) error {
	basePath := fmt.Sprintf("./%d", rand.Int())
	file, err := ai.TTS.ToFile(text, basePath)
	if err != nil {
		return fmt.Errorf("got err when creating tts %v\n", err)
	}
	defer os.Remove(file)

	if ai.Discord.voiceConnection == nil {
		return fmt.Errorf("voice connection is nil, can't say anything")
	}
	dgvoice.PlayAudioFile(ai.Discord.voiceConnection, file, make(chan bool))
	ai.Discord.session.UpdateGameStatus(0, fmt.Sprintf("responding with %s", text))
	defer ai.Discord.session.UpdateGameStatus(0, "")

	return nil
}

type CoquiConfig struct {
	Address string
	Voice   string
}

type ElevenLabsConfig struct {
	APIKey string
	Voice  string
}

type TTSConfig struct {
	Type       string
	Coqui      CoquiConfig
	ElevenLabs ElevenLabsConfig
}

func New(guildId string, discordSession *discordgo.Session, openAIClient *openai.Client, ttsConfig TTSConfig) *AI {
	var textToSpeech tts.TTS
	if ttsConfig.Type == "coqui" {
		textToSpeech = tts.NewCoqui(ttsConfig.Coqui.Address, ttsConfig.Coqui.Voice)
	}
	if ttsConfig.Type == "elevenlabs" {
		textToSpeech = tts.NewElevenlabs(ttsConfig.ElevenLabs.APIKey, ttsConfig.ElevenLabs.Voice)
	}

	return &AI{
		queue:        make([]Cmd, 0),
		openAIClient: openAIClient,
		Discord: &discordChannelSession{
			guildId: guildId,
			session: discordSession,
		},
		TTS: textToSpeech,
	}
}
