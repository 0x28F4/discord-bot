import queue
import time
import discord as dc

def get_current_time() -> int:
    return int(round(time.time() * 1000))


class _DiscordStream():
    """Opens a recording stream as a generator yielding the audio chunks."""

    def __init__(self) -> None:
        self.sample_width = 2
        self.channels = 2
        self.sample_rate = 48000

        self._buff: queue.Queue = queue.Queue()
        self.closed = False

    def write(self, data, user) -> None:
        # could check if the user is the author or not
        self._buff.put(data)

    def generator(self):
        """Stream Audio from discord to API and to local buffer

        returns:
            The data from the audio stream.
        """
        while not self.closed:
            data = []

            # Use a blocking get() to ensure there's at least one chunk of
            # data, and stop iteration if the chunk is None, indicating the
            # end of the audio stream.
            chunk = self._buff.get()
            if chunk is None:
                # do we ever get here???
                print("chunk is None, exiting generator")
                return
            data.append(chunk)
            while True:
                try:
                    chunk = self._buff.get(block=False)
                    if chunk is None:
                        return
                    data.append(chunk)

                except queue.Empty:
                    break

            yield b"".join(data)



class Sink(dc.sinks.Sink):
    """
    Need this class to capture audio buffers from the pycord library
    """

    def __init__(self, *, listen_to: int, filters=None):
        if filters is None:
            filters = dc.sinks.default_filters
        assert listen_to is not None
        self.filters = filters
        dc.sinks.Filters.__init__(self, **self.filters)
        self.vc = None
        self.audio_data = {}
        self.user_id = listen_to
        self.stream = _DiscordStream()

    def write(self, data, user):
        if user == self.user_id:
            self.stream.write(data=data, user=user)

    def cleanup(self):
        self.finished = True

    def get_all_audio(self):
        pass

    def get_user_audio(self, user):
        pass

    def set_user(self, user_id: int):
        self.user_id = user_id
        print(f"Set user ID: {user_id}")
