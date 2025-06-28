package tgbot

import (
	"fmt"
	"log"
	"os"
	"stroy-svaya/internal/model"
	bm "stroy-svaya/internal/tgbot/botmenu"
	"stroy-svaya/internal/tgbot/webservice"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type UserState struct {
	currRec    model.PileDrivingRecordLine
	currMenu   bm.DynamicMenu
	userName   string
	waitingFor string
}

type TgBot struct {
	bot        *tgbotapi.BotAPI
	userStates map[int64]*UserState
	ws         *webservice.WebService
}

func NewTgBot() *TgBot {
	ws := webservice.NewWebService("")
	return &TgBot{ws: ws}
}

func (b *TgBot) Run() error {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	tg_token := os.Getenv("TG_TOKEN")

	b.bot, err = tgbotapi.NewBotAPI(tg_token)
	if err != nil {
		log.Panic(err)
	}
	b.bot.Debug = true
	log.Printf("Authorized on account %s", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		msg := tgbotapi.NewMessage(chatID, "")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

		if _, ok := b.userStates[chatID]; !ok {
			b.userStates[chatID] = &UserState{}
		}
		if b.userStates[chatID].userName == "" {
			userName := fmt.Sprintf("%s %s",
				update.Message.From.FirstName,
				update.Message.From.LastName)
			b.userStates[chatID].userName, err = b.ws.GetUserFullName(chatID, userName)
			if err != nil {
				log.Println(err)
			}
		}

		switch text {
		case "/start":
			b.showPilesMenu(chatID)
		default:
			b.processUserInput(chatID, text)
		}
	}

	return nil
}

func (b *TgBot) showPilesMenu(chatID int64) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Все сваи", bm.PilesAll),
			tgbotapi.NewInlineKeyboardButtonData("Незабитые", bm.PilesNew),
			tgbotapi.NewInlineKeyboardButtonData("Без ФОГ", bm.PilesNoFPH),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Забитые вчера", bm.PilesLoggedYesterday),
			tgbotapi.NewInlineKeyboardButtonData("Забитые сегодня", bm.PilesLoggedToday),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Получить Excel", bm.PilesSendExcel),
		),
	)
	b.newInlineKb(chatID, &kb, "Добро пожаловать в журнал забивки свай!\n")
}

func (b *TgBot) showPileSelectionMenu(chatID int64, mode string) {
	usr := b.userStates[chatID]
	rec := model.PileDrivingRecordLine{}
	rec.ProjectId = 1
	rec.PileFieldId = 1
	rec.RecordedBy = usr.userName

	b.userStates[chatID].currRec = rec
	filter := model.PileFilter{}
	filter.ProjectId = rec.ProjectId
	filter.PileFieldId = rec.PileFieldId
	filter.RecordedBy = new(string)
	*filter.RecordedBy = rec.RecordedBy
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
		*filter.StartDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		filter.Status = 20
	case bm.PilesLoggedYesterday:
		now := time.Now()
		filter.StartDate = new(time.Time)
		*filter.StartDate = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
		filter.Status = 20
	}
	piles, err := b.ws.GetPiles(filter)
	if err != nil {
		log.Println(err)
	}

	if len(piles) == 0 {
		b.sendMessage(chatID, "Отсутствуют сваи заданным критериям")
		return
	}

	b.userStates[chatID].waitingFor = bm.PileSelection
	b.userStates[chatID].currMenu = *bm.NewDynamicMenu(piles)
	b.userStates[chatID].currMenu.BuildMenuOrHandleSelection(piles)
	// TO-DO НАРИСОВАТь МЕНЮ
}

func (b *TgBot) newInlineKb(chatID int64, kb *tgbotapi.InlineKeyboardMarkup, text string) {
	// msg := tgbotapi.NewMessage(chatID, "")
	// msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	// if _, err := b.bot.Send(msg); err != nil {
	// 	log.Panic(err)
	// }
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	if _, err := b.bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func (b *TgBot) processUserInput(chatID int64, text string) {
	switch text {
	case bm.PilesAll,
		bm.PilesNew,
		bm.PilesNoFPH,
		bm.PilesLoggedToday,
		bm.PilesLoggedYesterday:
		b.sendMessage(chatID, text)
	default:
		b.sendMessage(chatID, "other")
	}
	return
}

func (b *TgBot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}
}
