import queue
from typing import Optional
import discord as dc
from dataclasses import dataclass

@dataclass
class AudioData():
    silence: bool
    data: Optional[bytes] = None
        
class Sink(dc.sinks.Sink):
    """
    Need this class to capture audio buffers from the pycord library
    """

    def __init__(self, *, listen_to: int, silence_threshold_in_seconds = 1, filters=None):
        if filters is None:
            filters = dc.sinks.default_filters
        assert listen_to is not None
        self.filters = filters
        dc.sinks.Filters.__init__(self, **self.filters)
        self.vc = None
        self.audio_data = {}
        self.user_id = listen_to
        self._buff: queue.Queue = queue.Queue()
        self.silence_threshold_in_seconds = silence_threshold_in_seconds
        self.closed = False

    def write(self, data, user):
        if user == self.user_id:
            self._buff.put(data)

    def cleanup(self):
        self.finished = True

    def get_all_audio(self):
        pass

    def get_user_audio(self, user):
        pass

    def set_user(self, user_id: int):
        self.user_id = user_id
        print(f"Set user ID: {user_id}")

    def read(self):
        audio_received = False
        while not self.closed:
            try:
                data = []
                chunk = self._buff.get(block=True, timeout=self.silence_threshold_in_seconds)
                data.append(chunk)
                while True:
                    try:
                        chunk = self._buff.get(block=False)
                        if chunk is None:
                            return
                        data.append(chunk)
                    except queue.Empty:
                        audio_received = True
                        break
            except queue.Empty:
                if audio_received:
                    yield AudioData(silence=True, data=b"")
                    audio_received = False
                continue
            yield AudioData(silence=False, data=b"".join(data))
