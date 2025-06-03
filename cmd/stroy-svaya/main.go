package main

import (
	"log"
	"stroy-svaya/internal/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal("Failed to start app")
	}
	a.Run()
}
