package main

import (
	"discord-music-bot/configs"
	"discord-music-bot/services"
	"flag"
	"log"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(1)
	filename := flag.String("f", "bot.toml", "Set path for the config file.")
	flag.Parse()
	log.Println("INFO: Opening", *filename)
	err := configs.LoadConfig(*filename)
	if err != nil {
		log.Println("FATA:", err)
		return
	}
	// Hot reload
	configs.Watch()
	// Connecto to Discord
	err = services.DiscordConnect()
	if err != nil {
		log.Println("FATA: Discord", err)
		return
	}
	err = services.CreateDB()
	if err != nil {
		log.Println("FATA: DB", err)
		return
	}
	<-make(chan struct{})
}
