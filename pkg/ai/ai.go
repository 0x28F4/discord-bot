package ai

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/0x28F4/discord-bot/pkg/engines/decoder"
	"github.com/0x28F4/discord-bot/pkg/engines/voiceactivation"
	"github.com/0x28F4/discord-bot/pkg/prompts"
	"github.com/0x28F4/discord-bot/pkg/stt"
	"github.com/0x28F4/discord-bot/pkg/tts"
	"github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

type Cmd interface {
	Do(ai *AI) error
}

type AI struct {
	sttEngine    *voiceactivation.Engine
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

const (
	sampleRate  = engine.SampleRate // (16000)
	channels    = 1                 // decode into 1 channel since that is what whisper.cpp wants
	frameSizeMs = 20
)

var frameSize = channels * frameSizeMs * sampleRate / 1000

func heyGPT(text string) bool {
	// Ignore case using "(?i)" in the regex pattern.
	match, err := regexp.MatchString("(?i)(hey|hi|hello|cheers|hay).{0,5}(gpt|ai|waifu|google|siri|alexa)", text)
	if err != nil {
		return false
	}
	if match {
		return true
	}
	match, err = regexp.MatchString("(?i)(siri|alexa|dick)", text)
	if err != nil {
		return false
	}

	return match
}

func fuckoff(text string) bool {
	match, err := regexp.MatchString("(?i)fuck.{0,5}off", text)
	if err != nil {
		return false
	}
	return match
}

type ListenCmd struct{}

func (a *ListenCmd) Do(ai *AI) error {
	dec, err := decoder.NewOpusDecoder(sampleRate, channels)
	if err != nil {
		return err
	}

	var firstTimeStamp uint32
	pcm := make([]float32, frameSize)
	v := ai.Discord.voiceConnection
	go func() {
		for {
			if v.Ready == false || v.OpusRecv == nil {
				return
			}

			pkt, ok := <-v.OpusRecv
			if !ok {
				return
			}

			if _, err := dec.Decode(pkt.Opus, pcm); err != nil {
				return
			}

			if firstTimeStamp == 0 {
				fmt.Printf("Resetting timestamp bc firstTimeStamp is 0...  timestamp=%d\n", pkt.Timestamp)
				firstTimeStamp = pkt.Timestamp
			}

			ai.sttEngine.Write(pcm)
		}
	}()

	alreadyTalking := false
	ai.sttEngine.OnTranscribe = func(trans engine.Transcription) {
		text := make([]string, 0)
		for _, segment := range trans.Transcriptions {
			text = append(text, segment.Text)
		}

		prompt := strings.Join(text, " ")

		if fuckoff(prompt) {
			if err := ai.Discord.LeaveChannel(); err != nil {
				fmt.Printf("can't leave, %v\n", err)
			}
		}

		if heyGPT(prompt) {
			if alreadyTalking {
				fmt.Printf("already talking, skipping request")
				return
			}
			alreadyTalking = true
			defer func() {
				alreadyTalking = false
			}()
			toSay, err := ai.think(prompt, "")
			if err != nil {
				fmt.Printf("can't think, %v\n", err)
				return
			}

			if err := ai.say(toSay); err != nil {
				fmt.Printf("can't say, %v\n", err)
			}
		}
	}

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

	whisperCpp, err := stt.New("./models/ggml-base.en.bin")
	if err != nil {
		log.Fatalln(err, "error creating whisper model")
	}

	sttEngine, err := voiceactivation.New(voiceactivation.Params{
		Transcriber: whisperCpp,
	})

	return &AI{
		sttEngine:    sttEngine,
		queue:        make([]Cmd, 0),
		openAIClient: openAIClient,
		Discord: &discordChannelSession{
			guildId: guildId,
			session: discordSession,
		},
		TTS: textToSpeech,
	}
}
