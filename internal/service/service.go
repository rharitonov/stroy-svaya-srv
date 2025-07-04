package service

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"stroy-svaya/internal/model"
	"stroy-svaya/internal/repository"
	"time"

	"github.com/joho/godotenv"
	"github.com/tealeg/xlsx"
	"gopkg.in/gomail.v2"
)

type Service struct {
	repo *repository.SQLiteRepository
}

func NewService(r *repository.SQLiteRepository) *Service {
	return &Service{repo: r}
}

func (s *Service) InitPileDrivingRecordLine() *model.PileDrivingRecordLine {
	return &model.PileDrivingRecordLine{
		RecordedBy: "Фамилия Имя Отчество",
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

func (s *Service) SendPileDrivingRecordLog(projectId int) error {
	filename, err := s.SavePileDrivingRecordLogToExcel(projectId)
	if err != nil {
		return err
	}

	if err := s.SendMail(filename); err != nil {
		return err
	}

	return nil
}

func (s *Service) SavePileDrivingRecordLogToExcel(projectId int) (string, error) {
	var lines []model.PileDrivingRecordLine
	lines, err := s.GetPileDrivingRecord(projectId)
	if err != nil {
		return "", err
	}

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Журнал")
	if err != nil {
		panic(err)
	}

	row := sheet.AddRow()
	row.AddCell().Value = "Номер сваи"
	row.AddCell().Value = "Дата забивки"
	row.AddCell().Value = "Факт. отметка верха головы"
	row.AddCell().Value = "Оператор"
	for _, ln := range lines {
		row = sheet.AddRow()
		row.AddCell().Value = ln.PileNumber
		row.AddCell().Value = ln.StartDate.Format("02.01.2006")
		row.AddCell().Value = fmt.Sprintf("%d", ln.FactPileHead)
		row.AddCell().Value = ln.RecordedBy
	}

	printoutDate := time.Now()
	filename := fmt.Sprintf("./reports/p%d_журнал-забивки-свай-от_%s.xlsx",
		projectId,
		printoutDate.Format("2006-01-02_15-04-05"))
	err = file.Save(filename)
	if err != nil {
		panic(err)
	}
	return filename, nil
}

func (s *Service) SendMail(filename string) error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	sender := os.Getenv("MAIL_SENDER")
	password := os.Getenv("MAIL_SENDER_PASSWORD")
	to := os.Getenv("MAIL_TO")
	cc := os.Getenv("MAIL_CC")

	if sender == "" || password == "" || to == "" {
		return fmt.Errorf("missing required environment variables")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", to)
	if cc != "" {
		m.SetHeader("Cc", cc)
	}
	m.SetHeader("Subject", "=?UTF-8?B?"+base64.StdEncoding.EncodeToString([]byte("Журнал забивки свай"))+"?=")
	m.SetBody("text/plain", "см. вложение")

	if filename != "" {
		m.Attach(filename)
		defer os.Remove(filename)
	}

	d := gomail.NewDialer("smtp.yandex.ru", 465, sender, password)
	d.SSL = true

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Excel file %s sent, To %s, Cc %s", filename, to, cc)
	return nil
}

func (s *Service) GetPilesToDriving(projectId int) ([]string, error) {
	piles, err := s.repo.GetPilesToDriving(projectId)
	if err != nil {
		return nil, err
	}
	return piles, nil
}

func (s *Service) GetPiles(filter model.PileFilter) ([]string, error) {
	piles, err := s.repo.GetPiles(filter)
	if err != nil {
		return nil, err
	}
	return piles, nil
}

func (s *Service) GetPile(filter model.PileFilter) (*model.PileDrivingRecordLine, error) {
	pile, err := s.repo.GetPile(filter)
	if err != nil {
		return nil, err
	}
	return pile, nil
}

func (s *Service) InsertOrUpdatePdrPile(rec *model.PileDrivingRecordLine) error {
	return s.repo.InsertOrUpdatePdrPile(rec)
}

func (s *Service) GetUserFullNameInitialFormat(tgChatId int64) (string, error) {
	userName, err := s.repo.GetUserFullNameInitialFormat(tgChatId)
	if err != nil {
		return "", err
	}
	return userName, nil
}
