package service

import (
	"fmt"
	"stroy-svaya/internal/model"
	"stroy-svaya/internal/repository"
	"time"

	"github.com/tealeg/xlsx"
)

type Service struct {
	repo *repository.SQLiteRepository
}

func NewService(r *repository.SQLiteRepository) *Service {
	return &Service{repo: r}
}

func (s *Service) InitPileDrivingRecordLine() *model.PileDrivingRecordLine {
	return &model.PileDrivingRecordLine{
		RecordedBy: "Сатья Надела",
	}
}

func (s *Service) InsertPileDrivingRecordLine(rec *model.PileDrivingRecordLine) error {
	if err := s.repo.InsertPileDrivingRecordLine(rec); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetPileDrivingRecord(projectId int) ([]model.PileDrivingRecordLine, error) {

	lines, err := s.repo.GetPileDrivingRecord(projectId)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func (s *Service) PrintOutPileDrivingRecord(projectId int) error {
	var lines []model.PileDrivingRecordLine
	lines, err := s.GetPileDrivingRecord(projectId)
	if err != nil {
		return err
	}

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Журнал")
	if err != nil {
		panic(err)
	}

	row := sheet.AddRow()
	row.AddCell().Value = "Номер сваи"
	row.AddCell().Value = "Дата"
	row.AddCell().Value = "Отметка верха головы. факт"
	row.AddCell().Value = "Ответственный"
	for _, ln := range lines {
		row = sheet.AddRow()
		row.AddCell().Value = ln.PileNumber
		row.AddCell().Value = ln.StartDate.Format("02.01.2006")
		row.AddCell().Value = fmt.Sprintf("%d", ln.FactPileHead)
		row.AddCell().Value = ln.RecordedBy
	}

	printoutDate := time.Now()
	filename := fmt.Sprintf("./printout/p%d_журнал-забивки-свай-от_%s.xlsx",
		projectId,
		printoutDate.Format("2006-01-02_15-04-05"))
	err = file.Save(filename)
	if err != nil {
		panic(err)
	}
	return nil
}

func (s *Service) GetPilesToDriving(projectId int) ([]string, error) {

	piles, err := s.repo.GetPilesToDriving(projectId)
	if err != nil {
		return nil, err
	}
	return piles, nil
}
