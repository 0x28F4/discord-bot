import discord as dc

from discord_stream import DiscordStream


class Sink(dc.sinks.Sink):
    """
    Need this class to capture audio buffers from the pycord library
    """

    def __init__(self, *, stream: DiscordStream, listen_to: int, filters=None):
        if filters is None:
            filters = dc.sinks.default_filters
        assert listen_to is not None
        self.filters = filters
        dc.sinks.Filters.__init__(self, **self.filters)
        self.vc = None
        self.audio_data = {}
        self.user_id = listen_to
        self.stream = stream

    def write(self, data, user):
        if user == self.user_id:
            self.stream.write(data=data, user=user)

    def cleanup(self):
        self.finished = True

    def get_all_audio(self):
        pass

    def get_user_audio(self, user):
        pass

    # needed?
    def set_user(self, user_id: int):
        self.user_id = user_id
        print(f"Set user ID: {user_id}")
