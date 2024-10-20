import discord as dc
import speech_recognition

recognizer = speech_recognition.Recognizer()

class RecognizeSink(dc.sinks.Sink):
    def __init__(self, *, filters=None):
        if filters is None:
            filters = dc.sinks.default_filters
        self.filters = filters
        dc.sinks.Filters.__init__(self, **self.filters)
        self.vc = None
        self.audio_data = {}

        # user id for parsing their specific audio data
        self.user_id = None

        # obj to store our super sweet awesome audio data
        self.buffer = StreamBuffer()

    def write(self, data, user):
        # we overload the write method to take advantage of the already running thread for recording

        # if the data comes from the inviting user, we append it to buffer
        if user == self.user_id:
            self.buffer.write(data=data, user=user)

    def cleanup(self):
        self.finished = True

    def get_all_audio(self):
        # not applicable for streaming but may cause errors if not overloaded
        pass

    def get_user_audio(self, user):
        # not applicable for streaming but will def cause errors if not overloaded called
        pass

    def set_user(self, user_id: int):
        self.user_id = user_id
        print(f"Set user ID: {user_id}")


class StreamBuffer:
    def __init__(self) -> None:
        # holds byte-form audio data as it builds
        self.byte_buffer = bytearray()  # bytes

        # audio data specifications
        self.sample_width = 2
        self.channels = 2
        self.sample_rate = 48000
        self.bytes_ps = 192000  # bytes added to buffer per second
        self.block_len = 2  # how long you want each audio block to be in seconds
        # min len to pull bytes from buffer
        self.buff_lim = self.bytes_ps * self.block_len

        # var for tracking order of exported audio
        self.ct = 1

    # will need 'user' param if tracking multiple peoples voices - TBD
    def write(self, data, user) -> None:
        self.byte_buffer += data  # data is a bytearray object