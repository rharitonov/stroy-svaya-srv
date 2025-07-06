package main

import (
	"stroy-svaya/internal/config"
	"stroy-svaya/internal/repository"
	"stroy-svaya/internal/service"
)

func main() {
	c := config.Load()
	r, _ := repository.NewSQLiteRepository(c.DatabasePath)
	s := service.NewService(r)

	err := s.SendPileDrivingRecordLog(1, 204729745)
	if err != nil {
		panic(err)
	}
}
