package main

import (
	"log"
	"stroy-svaya/internal/app/tgbot"
)

func main() {
	bot := tgbot.NewTgBot()
	if err := bot.Run(); err != nil {
		log.Fatal("TG Bot can't start!")
	}
}
