package main

import (
	"log"
	"stroy-svaya/internal/config"
	"stroy-svaya/internal/repository"
	"stroy-svaya/internal/service"
)

func main() {
	c := config.Load()
	r, _ := repository.NewSQLiteRepository(c.DatabasePath)
	s := service.NewService(r)

	f, err := s.ExportPdrToExcel(1)
	if err != nil {
		panic(err)
	}
	log.Println(f)

}
