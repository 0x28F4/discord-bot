package ai

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type discordChannelSession struct {
	guildId   string
	channelId string

	session         *discordgo.Session
	voiceConnection *discordgo.VoiceConnection
}

func (d *discordChannelSession) JoinChannel(channelId string) error {
	voiceConnection, err := d.session.ChannelVoiceJoin(d.guildId, channelId, false, false)
	if err != nil {
		return fmt.Errorf("couldn't join discord channel, reason: %v", err)
	}

	d.voiceConnection = voiceConnection
	d.channelId = channelId

	return nil
}

func (d *discordChannelSession) LeaveChannel() error {
	if d.voiceConnection == nil {
		return nil
	}

	if err := d.voiceConnection.Disconnect(); err != nil {
		return fmt.Errorf("couldn't leave current voice connection, reason: %v", err)
	}

	d.voiceConnection = nil
	d.channelId = ""

	return nil
}

func (d *discordChannelSession) Connected() bool {
	return d.voiceConnection != nil
}
