package main

import (
	"fmt"
	"log"
	"os"
	"stroy-svaya/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate up|down")
	}
	cfg := config.Load()

	action := os.Args[1]
	m, err := migrate.New(
		"file://db/migrations",
		cfg.DatabaseUrl,
	)

	if err != nil {
		log.Fatalf("migrate error: %v", err)
	}

	switch action {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration applying error: %v", err)
		}
		fmt.Println("✅ migration applied")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration rollback error: %v", err)
		}
		fmt.Println("🔄 migration rolled back")

	default:
		log.Fatal("Unknown parameter:", action)
	}
}
