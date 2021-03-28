FROM golang:alpine
MAINTAINER Yurij Karpov <ljgago@gmail.com>

RUN apk update && apk add git ffmpeg ca-certificates && update-ca-certificates

RUN CGO_ENABLED=0 go get github.com/acrossoffwest/discord-music-bot

RUN mkdir /bot

WORKDIR /bot

CMD ["discord-music-bot", "-f", "bot.toml"]
