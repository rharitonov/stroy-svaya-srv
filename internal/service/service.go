package service

import (
	"encoding/base64"
	"fmt"
	"os"
	"stroy-svaya/internal/model"
	"stroy-svaya/internal/repository"
	"time"

	"github.com/joho/godotenv"
	"github.com/tealeg/xlsx"
	"github.com/xuri/excelize/v2"
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

func (s *Service) SendPileDrivingRecordLog(projectId int, tgChatId int64) error {
	u, err := s.GetUserSetup(tgChatId)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("send pdr excel file error, user %d not found", tgChatId)
	}
	if u.Email == "" {
		return fmt.Errorf("send pdr excel file error, user %d has no email", tgChatId)
	}
	filename, err := s.SavePileDrivingRecordLogToExcel(projectId)
	if err != nil {
		return err
	}

	if err := s.SendMail(u.Email, filename); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetProject(projectId int64) (*model.Project, error) {
	p, err := s.repo.GetProject(projectId)
	if err != nil {
		panic(err)
	}
	return p, nil
}

func (s *Service) ExportPdrToExcel(projectId int) (string, error) {

	f, err := excelize.OpenFile("templates/xlsx/pdr.xlsx")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sheet := "Журнал"

	{
		company_info, err := s.repo.GetCompanyName()
		if err != nil {
			panic(fmt.Errorf("company name: %w", err))
		}
		f.SetCellValue(sheet, "C1", company_info)
	}

	// project
	{
		p, err := s.repo.GetProject(int64(projectId))
		if err != nil {
			panic(err)
		}
		f.SetCellValue(sheet, "C3", p.Name)
		f.SetCellValue(sheet, "C5", p.Address)
		f.SetCellValue(sheet, "P14", p.StartDate)
		f.SetCellValue(sheet, "C18", p.StartDate)
		f.SetCellValue(sheet, "P15", p.EndDate)
		f.SetCellValue(sheet, "H18", p.EndDate)
	}

	equip := make(map[string]model.Equip)
	{
		equips, err := s.repo.GetPdrEquip(int64(projectId))
		if err != nil {
			panic(err)
		}
		var kopr, hummer, weight, power string
		for _, e := range equips {
			equip[e.Code] = e
			if len(kopr) != 0 {
				kopr += ","
			}
			kopr += e.Description
			if len(hummer) != 0 {
				hummer += ","
			}
			hummer += e.UnitType
			if len(weight) != 0 {
				weight += ","
			}
			weight += fmt.Sprintf("%d", e.UnitWeight)
			if len(power) != 0 {
				power += ","
			}
			power += fmt.Sprintf("%d", e.UnitPower)
		}

		f.SetCellValue(sheet, "F20", kopr)
		f.SetCellValue(sheet, "F21", hummer)
		f.SetCellValue(sheet, "F22", weight)
		f.SetCellValue(sheet, "F23", power)
	}

	printoutDate := time.Now()
	filename := fmt.Sprintf("reports/p%d_журнал-забивки-свай-от_%s.xlsx",
		projectId,
		printoutDate.Format("20060102_150405"))
	if err := f.SaveAs(filename); err != nil {
		panic(err)
	}
	return filename, nil
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
	if _, err := os.Stat("reports"); os.IsNotExist(err) {
		err := os.Mkdir("reports", 0755)
		if err != nil {
			panic(err)
		}
	}
	printoutDate := time.Now()
	filename := fmt.Sprintf("reports/p%d_журнал-забивки-свай-от_%s.xlsx",
		projectId,
		printoutDate.Format("20060102_150405"))
	err = file.Save(filename)
	if err != nil {
		panic(err)
	}
	return filename, nil
}

func (s *Service) SendMail(email string, filename string) error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	sender := os.Getenv("MAIL_SENDER")
	password := os.Getenv("MAIL_SENDER_PASSWORD")
	if sender == "" || password == "" || email == "" {
		return fmt.Errorf("missing required environment variables")
	}
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", email)
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
	//log.Printf("sending excel file %s, to %s", filename, email) //DBG
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

func (s *Service) GetUserSetup(tgChatID int64) (*model.User, error) {
	u, err := s.repo.GetUserSetup(tgChatID)
	if err != nil {
		return nil, err
	}
	return u, nil
}
