package main

import (
	"fmt"
	"stroy-svaya/internal/model"
	"stroy-svaya/internal/tgbot/webservice"
)

func main() {
	ws := webservice.NewWebService("")
	p0 := model.PileDrivingRecordLine{}
	p0.ProjectId = 1
	p0.PileFieldId = 1
	p0.PileNumber = "299"

	filter := model.PileFilter{}
	filter.ProjectId = p0.ProjectId
	filter.PileFieldId = p0.PileFieldId
	filter.PileNumber = &p0.PileNumber
	p, err := ws.GetPile(filter)
	if err != nil {
		panic(err)
	}
	fmt.Println(p)

	// c := config.Load()
	// r, _ := repository.NewSQLiteRepository(c.DatabasePath)
	// s := service.NewService(r)

	// filter := model.PileFilter{}
	// filter.ProjectId = 1
	// filter.PileFieldId = 1
	// filter.PileNumber = new(string)
	// *filter.PileNumber = "298"

	// p, err := s.GetPile(filter)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(*p)
}
