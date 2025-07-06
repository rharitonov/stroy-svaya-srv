package main

import (
	"log"
	"stroy-svaya/internal/tgbot/webservice"
)

func main() {
	w := webservice.NewWebService("")
	err := w.SendPdrLog(1, 201721111)

	if err != nil {
		log.Panic(err)
	}
}
