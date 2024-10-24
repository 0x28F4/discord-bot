FROM python:3.10
RUN apt update && apt install -y  ffmpeg

COPY src src
ENTRYPOINT ["/src/main.py"]
