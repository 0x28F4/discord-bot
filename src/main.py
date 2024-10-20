#!/usr/bin/env python3

from typing import Dict, cast
import discord as dc
import os
import io

from sink import RecognizeSink

assert (TOKEN := os.getenv('DISCORD_TOKEN'))

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
    force: dc.Option(dc.SlashCommandOptionType.boolean) # type: ignore
):
    for vc in connections.values():
        await vc.disconnect(force=force)
    await ctx.respond("Quit all VCs")

connections: Dict[int, dc.VoiceClient] = {}

async def join_channel(
    ctx: dc.ApplicationContext,
    channel: dc.channel.VoiceChannel,
):
    vc = await channel.connect()
    connections.update({ctx.guild.id: vc})

    def on_done(sink: RecognizeSink, channel: dc.TextChannel, *args):
        for user_id, audio_data in sink.audio_data.items():
            data = cast(dc.sinks.core.AudioData, audio_data)
            with open(f"./raw_{user_id}.out", "wb") as f:
                audio_file = cast(io.BytesIO, data.file)
                f.write(audio_file.getbuffer())
        
    vc.start_recording(
        RecognizeSink(),
        on_done,
        ctx.channel
    )
    await ctx.respond("Started recognizing!")

bot.run(TOKEN)