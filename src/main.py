#!/usr/bin/env python3

import asyncio
from typing import Dict, cast
import discord as dc
import os
from google.cloud.speech_v2 import SpeechClient
from google.cloud.speech_v2.types import cloud_speech as cloud_speech_types

from discord_stream import DiscordStream, listen_print_loop
from sink import Sink

assert (TOKEN := os.getenv('DISCORD_TOKEN'))
assert (GCP_PROJECT_ID := os.getenv("DISCORD_BOT_GCP_ID"))
SPEECH_MAX_CHUNK_SIZE = 25600

bot = dc.Bot()

@bot.event
async def on_ready():
    print(f"{bot.user} is ready and online!")

@bot.slash_command(name="hello", description="Say hello to the bot")
async def hello(ctx: dc.ApplicationContext):
    await ctx.respond("Hey!")

@bot.slash_command(name="follow", description="Join my vc and recoqnize speech")
async def follow(
    ctx: dc.ApplicationContext,
):
    author = ctx.author
    if 'voice' not in dir(author):
        await ctx.respond("Must be called in guild")
        return

    voice = author.voice
    if not voice:
        await ctx.respond("You aren't in a voice channel!")
        return
    
    await join_channel(ctx, cast(dc.VoiceState, voice).channel)

@bot.slash_command(name="join", description="Join vc and recognize speech")
async def join(
    ctx: dc.ApplicationContext,
    channel: dc.Option(dc.SlashCommandOptionType.channel) # type: ignore
):
    await join_channel(ctx, channel)

@bot.slash_command(name="quit", description="Quits all vcs")
async def quit(
    ctx: dc.ApplicationContext,
):
    for vc in connections.values():
        await vc.disconnect()
    await ctx.respond("Quit all VCs")

connections: Dict[int, dc.VoiceClient] = {}

async def join_channel(
    ctx: dc.ApplicationContext,
    channel: dc.channel.VoiceChannel,
):
    client = SpeechClient()
    recognition_config = cloud_speech_types.RecognitionConfig(
        explicit_decoding_config=cloud_speech_types.ExplicitDecodingConfig(
            sample_rate_hertz=48000,
            encoding=cloud_speech_types.ExplicitDecodingConfig.AudioEncoding.LINEAR16,
            audio_channel_count=2
        ),
        language_codes=["en-US"],
        model="long",
    )
    streaming_config = cloud_speech_types.StreamingRecognitionConfig(
        config=recognition_config,
        streaming_features=cloud_speech_types.StreamingRecognitionFeatures(
            interim_results=True
        )
    )
    config_request = cloud_speech_types.StreamingRecognizeRequest(
        recognizer=f"projects/{GCP_PROJECT_ID}/locations/global/recognizers/_",
        streaming_config=streaming_config,
    )
    
    vc = await channel.connect()
    connections.update({ctx.guild.id: vc})

    async def on_done(sink: Sink, channel: dc.TextChannel, *args):
        pass

    stream=DiscordStream()
    sink = Sink(stream=stream)
    sink.user_id = ctx.author.id


    def requests_gen(config: cloud_speech_types.RecognitionConfig, audio: list):
        """Helper function to generate the requests list for the streaming API.

        Args:
            config: The speech recognition configuration.
            audio: The audio data.
        Returns:
            The list of requests for the streaming API.
        """
        yield config
        for buffer in audio:
            for i in range(0, len(buffer), SPEECH_MAX_CHUNK_SIZE):
                chunk = buffer[i:SPEECH_MAX_CHUNK_SIZE]
                if any(chunk):
                    yield cloud_speech_types.StreamingRecognizeRequest(audio=chunk)


    audio_generator = stream.generator()
    vc.start_recording(
        sink,
        on_done,
        ctx.channel
    )
    await ctx.respond("Started recognizing!")
    responses_iterator = client.streaming_recognize(requests=requests_gen(config_request, audio_generator))
    listen_print_loop(responses_iterator, stream)

bot.run(TOKEN)
