FROM golang:1.20 as base
RUN apt update && apt install -y libopus-dev libopusfile-dev

FROM base as build-env
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . .
RUN go build -o discordbot . 

FROM base
COPY --from=build-env /app/discordbot /discordbot
ENTRYPOINT ["/discordbot"]