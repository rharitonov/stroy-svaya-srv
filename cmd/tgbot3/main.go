package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"stroy-svaya/internal/model"
	dynmenu "stroy-svaya/internal/tgbot/botmenu"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

const (
	WebServiceURL = "http://localhost:8080"
	DateFormat    = "02.01.2006"
	GroupsCount   = 6
	HelpTxt       = "Используйте команды:\n" +
		"/newrecord - начать новую запись\n" +
		"/sendexcel - отправить excel c данными на рабочую почту\n" +
		"/help - помощь"
)

type UserState struct {
	WaitingFor       string
	CurrentRecord    model.PileDrivingRecordLine
	AvailablePiles   []string
	SelectionHistory [][]string
	DynamicMenu      *dynmenu.DynamicMenu
}

var (
	bot        *tgbotapi.BotAPI
	userStates = make(map[int64]*UserState)
)

func main() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	tg_token := os.Getenv("TG_TOKEN")

	bot, err = tgbotapi.NewBotAPI(tg_token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		userName := fmt.Sprintf("%s %s",
			update.Message.From.FirstName,
			update.Message.From.LastName)
		userName = getUserFullName(chatID, userName)

		if _, ok := userStates[chatID]; !ok {
			userStates[chatID] = &UserState{}
		}

		state := userStates[chatID]

		msg := tgbotapi.NewMessage(chatID, "")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

		switch text {
		case "/start":
			sendMessage(chatID, "Добро пожаловать в журнал забивки свай!\n\n"+HelpTxt)
		case "/help":
			sendMessage(chatID, HelpTxt)
		case "/newrecord":
			startNewRecord(chatID, userName, state)
		case "/sendexcel":
			sendPileDrivingLog(chatID, state)
		default:
			processUserInput(chatID, state, text)
		}
	}
}

func startNewRecord(chatID int64, userName string, state *UserState) {
	state.CurrentRecord = model.PileDrivingRecordLine{
		ProjectId:   1,
		PileFieldId: 1,
		RecordedBy:  userName,
	}

	piles := getPilesToDriving()
	if len(piles) == 0 {
		sendMessage(chatID, "Нет доступных свай для забивки.")
		return
	}
	state.WaitingFor = "pileNumber"

	state.DynamicMenu = dynmenu.NewDynamicMenu(piles)
	state.DynamicMenu.BuildMenuOrHandleSelection(piles)
	msg := tgbotapi.NewMessage(chatID, "Выберите группу свай:")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(state.DynamicMenu.TgBotMenu...)
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}
}

func sendPileDrivingLog(chatID int64, state *UserState) {
	url := fmt.Sprintf("%s/sendpdrlog?project_id=1", WebServiceURL)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer resp.Body.Close()
	sendMessage(chatID, "Excel файл отправлен.")
	state.WaitingFor = ""
	state.SelectionHistory = [][]string{}
	sendMessage(chatID, HelpTxt)
}

func getPilesToDriving() []string {
	url := fmt.Sprintf("%s/getpilestodriving?project_id=1", WebServiceURL)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	var piles []string
	if err := json.Unmarshal(body, &piles); err != nil {
		log.Fatal(err.Error())
	}
	return piles
}

func getUserFullName(chatId int64, defaultUserName string) string {
	url := fmt.Sprintf("%s/getuserfullname?tg_chat_id=%d", WebServiceURL, chatId)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	var userName string
	if err := json.Unmarshal(body, &userName); err != nil {
		log.Fatal(err.Error())
	}
	if userName == "" {
		userName = defaultUserName
	}
	return userName
}

func processUserInput(chatID int64, state *UserState, text string) {
	switch state.WaitingFor {
	case "pileNumber":
		handlePileNumberSelection(chatID, state, text)
	case "drivingDate":
		handleDrivingDateSelection(chatID, state, text)
	case "pileTopLevel":
		handlePileTopLevelInput(chatID, state, text)
	default:
		sendMessage(chatID, HelpTxt)
	}
}

func handlePileNumberSelection(chatID int64, state *UserState, text string) {
	state.DynamicMenu.BuildMenuOrHandleSelection(text)
	if state.DynamicMenu.SingleItemSelected() {
		state.CurrentRecord.PileNumber = text
		state.WaitingFor = "fin"
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Выбрана свая: %s", state.CurrentRecord.PileNumber))
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, err := bot.Send(msg)
		if err != nil {
			log.Println("Ошибка при отправке сообщения:", err)
		}

	} else {
		msg := tgbotapi.NewMessage(chatID, "Выберите группу свай:")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(state.DynamicMenu.TgBotMenu...)
		_, err := bot.Send(msg)
		if err != nil {
			log.Println("Ошибка при отправке сообщения:", err)
		}
	}
}

func sendDateSelection(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Выберите дату забивки:")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Сегодня"),
			tgbotapi.NewKeyboardButton("Вчера"),
		),
	)

	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}
}

func handleDrivingDateSelection(chatID int64, state *UserState, text string) {
	now := time.Now()
	var selectedDate time.Time

	switch text {
	case "Сегодня":
		selectedDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case "Вчера":
		selectedDate = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
	default:
		sendMessage(chatID, "Пожалуйста, выберите одну из предложенных дат")
		sendDateSelection(chatID)
		return
	}

	state.CurrentRecord.StartDate = selectedDate
	state.WaitingFor = "pileTopLevel"
	sendMessage(chatID, fmt.Sprintf("Выбрана дата: %s\nВведите отметку верха головы сваи (в милиметрах, например, 10750):",
		selectedDate.Format(DateFormat)))
}

func handlePileTopLevelInput(chatID int64, state *UserState, text string) {
	factPileHead, err := parseInt(text)
	if err != nil {
		sendMessage(chatID, "Неверный формат числа. Пожалуйста, введите отметку в милиметрах (например, 10750):")
		return
	}
	state.CurrentRecord.FactPileHead = factPileHead
	sendDataToWebService(chatID, state)
}

func sendDataToWebService(chatID int64, state *UserState) {
	jsonData, err := json.Marshal(state.CurrentRecord)
	if err != nil {
		sendMessage(chatID, "Ошибка при подготовке данных: "+err.Error())
		return
	}

	url := fmt.Sprintf("%s/insertpdrline", WebServiceURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		sendMessage(chatID, "Ошибка при отправке данных на сервер: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// Убираем клавиатуру после отправки
	msg := tgbotapi.NewMessage(chatID, "")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	if resp.StatusCode == http.StatusCreated {
		msg.Text = "Данные успешно отправлены!\n\n" +
			"Номер сваи: " + state.CurrentRecord.PileNumber + "\n" +
			"Дата забивки: " + state.CurrentRecord.StartDate.Format(DateFormat) + "\n" +
			"Отметка верха: " + fmt.Sprintf("%d", state.CurrentRecord.FactPileHead) + " мм"
	} else {
		msg.Text = "Сервер вернул ошибку: " + resp.Status
	}

	_, err = bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}

	// Сбрасываем состояние пользователя
	state.WaitingFor = ""
	state.SelectionHistory = [][]string{}
	sendMessage(chatID, HelpTxt)
}

func sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
