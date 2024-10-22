import struct

def pcm_to_float(pcm_data):
    integer_value = struct.unpack('<h', pcm_data)[0]
    float_value = integer_value / 32768.0
    return float_value
