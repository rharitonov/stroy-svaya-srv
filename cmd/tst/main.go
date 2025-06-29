package main

import (
	"fmt"
	"log"
	"stroy-svaya/internal/model"
	bm "stroy-svaya/internal/tgbot/botmenu"
	"stroy-svaya/internal/tgbot/webservice"
	"time"
)

func main() {
	//c := config.Load()
	//r, _ := repository.NewSQLiteRepository(c.DatabasePath)
	//s := service.NewService(r)

	filter := model.PileFilter{}
	filter.ProjectId = 1
	filter.PileFieldId = 1
	filter.RecordedBy = new(string)
	*filter.RecordedBy = "Вася"

	mode := bm.PilesAll
	// if len(os.Args) != 2 {
	// 	//log.Fatal("mode reqiured")
	// }
	// mode := os.Args[1]
	switch mode {
	case bm.PilesAll:
		filter.Status = 30
	case bm.PilesNew:
		filter.Status = 10
	case bm.PilesNoFPH:
		filter.FactPileHead = new(int)
		*filter.FactPileHead = 0
		filter.Status = 20
	case bm.PilesLoggedToday:
		now := time.Now()
		filter.StartDate = new(time.Time)
		//*filter.StartDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		*filter.StartDate = time.Date(now.Year(), now.Month(), 28, 0, 0, 0, 0, time.UTC)
		filter.Status = 20
	case bm.PilesLoggedYesterday:
		now := time.Now()
		filter.StartDate = new(time.Time)
		//*filter.StartDate = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
		*filter.StartDate = time.Date(now.Year(), now.Month(), 27, 0, 0, 0, 0, time.UTC)
		filter.Status = 20
	default:
		log.Fatal("incorrect mode value")
	}

	w := webservice.NewWebService("")
	piles, err := w.GetPiles(filter)
	//piles, err := s.GetPiles(filter)
	if err != nil {
		log.Panic(err)
	}
	if len(piles) == 0 {
		log.Panic("Отсутствуют сваи заданным критериям")
	}

	fmt.Printf("for %s:\n%v\n\n", mode, piles)
}
