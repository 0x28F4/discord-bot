import queue
import time


def get_current_time() -> int:
    return int(round(time.time() * 1000))


RED = "\033[0;31m"
GREEN = "\033[0;32m"
YELLOW = "\033[0;33m"


class DiscordStream:
    """Opens a recording stream as a generator yielding the audio chunks."""

    def __init__(self) -> None:
        self.sample_width = 2
        self.channels = 2
        self.sample_rate = 48000

        self._buff = queue.Queue()
        # self.closed = True
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
