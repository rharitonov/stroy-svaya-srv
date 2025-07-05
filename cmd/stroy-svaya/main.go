package main

import (
	"log"
	app "stroy-svaya/internal/app/stroy-svaya"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal("Failed to start app")
	}
	a.Run()
}
