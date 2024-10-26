#!/usr/bin/env python3

import io
from typing import Dict, cast
import discord as dc
import os

from chat import Chat
import config
from discord_audio import Sink
from engine import Engine
from stt import STT
from tts import TTS
from utils import DEBUG

assert (TOKEN := os.getenv("DISCORD_TOKEN"))

bot = dc.Bot()
cfg = config.load()

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
    if not hasattr(author, 'voice'):
        await ctx.respond("Must be called in guild")
        return
    voice = author.voice
    if not voice:
        await ctx.respond("You aren't in a voice channel!")
        return

    channel = cast(dc.channel.VoiceChannel, cast(dc.VoiceState, voice).channel) 
    await join_channel(ctx, channel, author.id)


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


def noop(*args, **kwargs): pass

async def join_channel(
    ctx: dc.ApplicationContext,
    channel: dc.channel.VoiceChannel,
    listen_to: int,
):
    vc: dc.VoiceClient = await connect_channel(ctx, channel)
    sink = Sink(listen_to=listen_to)
    vc.start_recording(sink, callback=noop)
    def handle_audio(audio_data: bytes):
        if vc.is_playing():
            if DEBUG(): print("skip handling audio, because already playing")
            return
        buffer = io.BytesIO(audio_data)
        buffer.seek(0)
        source = dc.FFmpegPCMAudio(buffer, pipe=True)
        vc.play(source)

    chat = Chat(config=cfg["chat"])
    tts = TTS(config=cfg["tts"])
    stt = STT()
    engine = Engine(chat=chat, tts=tts, on_audio=handle_audio)
    await ctx.respond("Started recognizing!")

    import threading
    def speech_loop():
        try:
            responses = stt.stream(source=sink.stream.generator())
            engine.run(responses=responses, user=ctx.author.name)
        except Exception as e:
            print("got exception in speech loop: ", e)

    t = threading.Thread(
        target=speech_loop,
    )
    t.start()
    print("[join] end")


bot.run(TOKEN)
