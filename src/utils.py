import os


def DEBUG():
    return os.getenv("DEBUG") != ""
