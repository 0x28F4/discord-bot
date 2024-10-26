#!/usr/bin/env python3

import io
from typing import Dict, cast
import discord as dc
import os
from google.cloud.speech_v2 import SpeechClient
from google.cloud.speech_v2.types import cloud_speech as cloud_speech_types

from chat import Chat
import config
from discord_stream import DiscordStream
from brain import respond
from sink import Sink
from utils import DEBUG

assert (TOKEN := os.getenv("DISCORD_TOKEN"))
assert (GCP_PROJECT_ID := os.getenv("DISCORD_BOT_GCP_ID"))
SPEECH_MAX_CHUNK_SIZE = 25600

bot = dc.Bot()
chat = Chat(config=config.load())


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
    if "voice" not in dir(author):
        await ctx.respond("Must be called in guild")
        return

    voice = author.voice
    if not voice:
        await ctx.respond("You aren't in a voice channel!")
        return

    await join_channel(ctx, cast(dc.VoiceState, voice).channel, author.id)


@bot.slash_command(name="join", description="Join vc and recognize speech")
async def join(
    ctx: dc.ApplicationContext,
    channel: dc.Option(dc.SlashCommandOptionType.channel),  # type: ignore
    user: dc.Option(dc.SlashCommandOptionType.user),  # type: ignore
):
    await join_channel(ctx, channel, user.id)


@bot.slash_command(name="quit", description="Quits all vcs")
async def quit(
    ctx: dc.ApplicationContext,
):
    for vc in connections.values():
        await vc.disconnect()
    await ctx.respond("Quit all VCs")


connections: Dict[int, dc.VoiceClient] = {}


async def connect_channel(ctx: dc.ApplicationContext, channel: dc.channel.VoiceChannel):
    vc = ctx.voice_client
    if vc:
        if vc.channel.id == channel.id:
            return vc

        # stop recording in other channel and disconnect
        if vc.recording:
            vc.stop_recording()
        await vc.disconnect()
    # anyways, gotta connect
    vc = await channel.connect()
    connections.update({ctx.guild.id: vc})
    return vc


async def join_channel(
    ctx: dc.ApplicationContext,
    channel: dc.channel.VoiceChannel,
    listen_to: int,
):
    vc: dc.VoiceClient = await connect_channel(ctx, channel)
    client = SpeechClient()
    recognition_config = cloud_speech_types.RecognitionConfig(
        explicit_decoding_config=cloud_speech_types.ExplicitDecodingConfig(
            sample_rate_hertz=48000,
            encoding=cloud_speech_types.ExplicitDecodingConfig.AudioEncoding.LINEAR16,
            audio_channel_count=2,
        ),
        language_codes=["en-US"],
        model="long",
    )
    streaming_config = cloud_speech_types.StreamingRecognitionConfig(
        config=recognition_config,
        streaming_features=cloud_speech_types.StreamingRecognitionFeatures(
            interim_results=True
        ),
    )
    config_request = cloud_speech_types.StreamingRecognizeRequest(
        recognizer=f"projects/{GCP_PROJECT_ID}/locations/global/recognizers/_",
        streaming_config=streaming_config,
    )

    stream = DiscordStream()
    sink = Sink(stream=stream, listen_to=listen_to)
    audio_generator = stream.generator()

    def requests_gen(config: cloud_speech_types.RecognitionConfig, audio: list):
        """Helper function to generate the requests list for the streaming API.

        Args:
            config: The speech recognition configuration.
            audio: The audio data.
        Returns:
            The list of requests for the streaming API.
        """

        started = False
        for buffer in audio:
            if not started:
                yield config
                started = True
            for i in range(0, len(buffer), SPEECH_MAX_CHUNK_SIZE):
                chunk = buffer[i:SPEECH_MAX_CHUNK_SIZE]
                # must turn off discords voice activation, otherwise this won't recognize silence
                if any(chunk):
                    yield cloud_speech_types.StreamingRecognizeRequest(audio=chunk)

    async def on_done(sink: Sink, channel: dc.TextChannel, *args):
        pass

    vc.start_recording(sink, on_done, ctx.channel)

    def handle_audio(audio_data: bytes):
        if vc.is_playing():
            if DEBUG(): print("skip handling audio, because already playing")
            return
        buffer = io.BytesIO(audio_data)
        buffer.seek(0)
        source = dc.FFmpegPCMAudio(buffer, pipe=True)
        vc.play(source)

    await ctx.respond("Started recognizing!")
    import threading

    def speech_loop(client, config_request, audio_generator):
        try:
            responses_iterator = client.streaming_recognize(
                requests=requests_gen(config_request, audio_generator)
            )
            respond(
                chat=chat,
                user=ctx.author.name,
                responses=responses_iterator,
                on_audio=handle_audio,
            )
        except Exception as e:
            print("got exception in speech loop: ", e)

    t = threading.Thread(
        target=speech_loop,
        args=(
            client,
            config_request,
            audio_generator,
        ),
    )
    t.start()
    print("[join] end")


bot.run(TOKEN)
