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
		log.Fatal("Использование: migrate up|down")
	}
	cfg := config.Load()

	action := os.Args[1]
	m, err := migrate.New(
		"file://db/migrations",
		cfg.DatabaseUrl,
	)

	if err != nil {
		log.Fatalf("Ошибка создания мигратора: %v", err)
	}

	switch action {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Ошибка применения миграции: %v", err)
		}
		fmt.Println("✅ Миграции применены")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Ошибка отката миграции: %v", err)
		}
		fmt.Println("🔄 Миграции откачены")

	default:
		log.Fatal("Неизвестная команда:", action)
	}
}
