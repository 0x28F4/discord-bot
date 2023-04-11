package ai

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

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
	TTS          *tts
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
}

func (a *AskCmd) Do(ai *AI) error {
	output, err := ai.think(a.Prompt)
	if err != nil {
		return fmt.Errorf("couldn't come up with an answer for %s\n", a.Prompt)
	} else if err := ai.say(output); err != nil {
		return fmt.Errorf("couldn't say \"%s\"\n", output)
	}

	return nil
}

func (ai *AI) think(prompt string) (out string, err error) {
	resp, err := ai.openAIClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "you are a discord bot and respond to messages by users",
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

	return resp.Choices[0].Message.Content, nil
}

func (ai *AI) say(text string) error {
	filepath := fmt.Sprintf("./%d.wav", rand.Int())
	if err := ai.TTS.ToFile(text, filepath); err != nil {
		return fmt.Errorf("got err when creating tts %v\n", err)
	}
	defer os.Remove(filepath)

	if ai.Discord.voiceConnection == nil {
		return fmt.Errorf("voice connection is nil, can't say anything")
	}
	dgvoice.PlayAudioFile(ai.Discord.voiceConnection, filepath, make(chan bool))
	ai.Discord.session.UpdateGameStatus(0, fmt.Sprintf("responding with %s", text))
	defer ai.Discord.session.UpdateGameStatus(0, "")

	return nil
}

func New(guildId, ttsAddress string, discordSession *discordgo.Session, openAIClient *openai.Client) *AI {
	return &AI{
		queue:        make([]Cmd, 0),
		openAIClient: openAIClient,
		Discord: &discordChannelSession{
			guildId: guildId,
			session: discordSession,
		},
		TTS: NewTTS(ttsAddress),
	}
}
