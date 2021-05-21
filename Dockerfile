FROM golang:alpine
MAINTAINER Yurij Karpov <ljgago@gmail.com>

WORKDIR /bot

RUN apk update && apk add git ffmpeg ca-certificates && update-ca-certificates

RUN git clone https://github.com/acrossoffwest/discord-music-bot.git /bot
RUN  cd /bot && CGO_ENABLED=0 go mod download

CMD cd /bot && go run .
