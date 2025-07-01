package tgbot

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
	tb := &TgBot{}
	tb.ws = webservice.NewWebService("")
	tb.userStates = make(map[int64]*UserState)
	return tb
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
		switch true {
		case update.Message != nil:
			chatID := update.Message.Chat.ID
			text := update.Message.Text
			b.getUserState(chatID, update.Message.From)
			b.processCommand(chatID, text)
		case update.CallbackQuery != nil:
			callback := update.CallbackQuery
			chatID := callback.Message.Chat.ID
			data := update.CallbackQuery.Data
			b.getUserState(chatID, callback.From)
			if data != bm.PilesSendExcel {
				b.processCallbackQuery(chatID, data)
			} else {
				b.processCallbackQueryWithAlert(chatID, data, callback)
			}
		default:
			continue
		}
	}
	return nil
}

func (b *TgBot) showPilesMenu(chatID int64) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Все сваи", bm.PilesAll),
			tgbotapi.NewInlineKeyboardButtonData("Незабитые", bm.PilesNew),
			tgbotapi.NewInlineKeyboardButtonData("Без ФОВГ", bm.PilesNoFPH),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Забитые сегодня", bm.PilesLoggedToday),
			tgbotapi.NewInlineKeyboardButtonData("Забитые вчера", bm.PilesLoggedYesterday),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Получить Excel", bm.PilesSendExcel),
		),
	)
	b.newInlineKb(chatID, &kb, "Добро пожаловать в журнал забивки свай!\n")
}

func (b *TgBot) showPileOperationsMenu(chatID int64) {
	var kb tgbotapi.InlineKeyboardMarkup
	if b.userStates[chatID].currRec.Status == 10 {
		kb = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Запись в журнал", bm.PileOpsInsert),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Ввод/изм. ФОВГ", bm.PileOpsUpdateFPH),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("В главное меню", bm.PileOpsBack),
			),
		)
	} else {
		kb = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Ввод/изм. ФОВГ", bm.PileOpsUpdateFPH),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("В главное меню", bm.PileOpsBack),
			),
		)
	}
	b.newInlineKb(chatID, &kb, "Доступные операции:")
}

func (b *TgBot) showPileStartDateMenu(chatID int64) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сегодня", bm.PileOpsStartDateToday),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вчера", bm.PileOpsStartDateYesterday),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("В главное меню", bm.PileOpsBack),
		),
	)
	b.newInlineKb(chatID, &kb, "Выберите дату забивки сваи:")
}

func (b *TgBot) showAfterUpdatePdrLineMenu(chatID int64) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("В главное меню", bm.PileOpsBack),
		),
	)
	b.newInlineKb(chatID, &kb, b.getPileInfo(chatID, "Данные успешно записаны в журнал:"))
}

func (b *TgBot) startPileSelection(chatID int64, mode string) {
	b.userStates[chatID].waitingFor = ""
	rec := model.PileDrivingRecordLine{}
	rec.ProjectId = 1
	rec.PileFieldId = 1
	rec.RecordedBy = b.userStates[chatID].userName

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
	b.userStates[chatID].currMenu = *bm.NewDynamicMenu(piles)
	b.makePileSelectionMenu(chatID, "")
	b.userStates[chatID].waitingFor = bm.WaitPileNumber
}

func (b *TgBot) newInlineKb(chatID int64, kb *tgbotapi.InlineKeyboardMarkup, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	if _, err := b.bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func (b *TgBot) processCommand(chatID int64, text string) {
	switch text {
	case "/start":
		b.showPilesMenu(chatID)
	default:
		b.processUserInput(chatID, text)
	}
}

func (b *TgBot) processUserInput(chatID int64, text string) {
	switch b.userStates[chatID].waitingFor {
	case bm.WaitPileUpdateFPH:
		b.onAfterPileUpdateFPH(chatID, text)
	default:
		b.sendMessage(chatID, fmt.Sprintf("Debug: %s, %s", b.userStates[chatID].waitingFor, text))
	}
}

func (b *TgBot) processCallbackQuery(chatID int64, data string) {
	switch data {
	case bm.PilesAll,
		bm.PilesNew,
		bm.PilesNoFPH,
		bm.PilesLoggedToday,
		bm.PilesLoggedYesterday:
		b.startPileSelection(chatID, data)
	case bm.PileOpsBack:
		b.showPilesMenu(chatID)
	default:
		switch b.userStates[chatID].waitingFor {
		case bm.WaitPileNumber:
			b.makePileSelectionMenu(chatID, data)
		case bm.WaitPileOperation:
			switch data {
			case bm.PileOpsUpdateFPH:
				b.onBeforePileUpdateFPH(chatID)
			case bm.PileOpsInsert:
				b.insertOrUpdatePile(chatID)
			}
		case bm.WaitPileUpdateFPH:
			b.onAfterPileUpdateFPH(chatID, data)
		case bm.WaitPileStartDate:
			b.onAfterStartDateSelect(chatID, data)
		}
	}
}

func (b *TgBot) processCallbackQueryWithAlert(chatID int64, data string, callback *tgbotapi.CallbackQuery) {
	switch data {
	case bm.PilesSendExcel:
		b.sendPdrLog(chatID, callback)
	}
}

func (b *TgBot) getUserState(chatID int64, tgUser *tgbotapi.User) {
	if _, ok := b.userStates[chatID]; !ok {
		b.userStates[chatID] = &UserState{}
	}
	if b.userStates[chatID].userName == "" {
		userName := fmt.Sprintf("%s %s",
			tgUser.FirstName,
			tgUser.LastName)
		var err error
		b.userStates[chatID].userName, err = b.ws.GetUserFullName(chatID, userName)
		if err != nil {
			panic(err)
		}
	}
}

func (b *TgBot) makePileSelectionMenu(chatID int64, data string) {
	if b.userStates[chatID].waitingFor == bm.WaitPileNumber {
		b.userStates[chatID].currMenu.BuildMenuOrHandleSelection(data)
	} else {
		b.userStates[chatID].currMenu.BuildMenuOrHandleSelection(nil)
	}
	if !b.userStates[chatID].currMenu.SingleItemSelected() {
		kb := b.userStates[chatID].currMenu.GetTgKeyboardMenu()
		b.newInlineKb(chatID, &kb, "Выберите номер сваи из предложенных вариантов:")
		return
	}
	b.updatePileRec(chatID, data)
	b.showPileInfo(chatID, "Выбрана свая:")
	b.showPileOperationsMenu(chatID)
	b.userStates[chatID].waitingFor = bm.WaitPileOperation
}

func (b *TgBot) updatePileRec(chatID int64, pile_no string) {
	b.userStates[chatID].currRec.PileNumber = pile_no
	filter := model.PileFilter{}
	filter.ProjectId = b.userStates[chatID].currRec.ProjectId
	filter.PileFieldId = b.userStates[chatID].currRec.PileFieldId
	filter.PileNumber = &b.userStates[chatID].currRec.PileNumber
	p, err := b.ws.GetPile(filter)
	if err != nil {
		panic(err)
	}
	b.userStates[chatID].currRec = *p
}

func (b *TgBot) getPileInfo(chatID int64, title string) string {
	p := b.userStates[chatID].currRec
	infoText := ""
	switch p.Status {
	case 10:
		infoText = "Номер сваи: %s\nСтатус: план"
		infoText = fmt.Sprintf(infoText, p.PileNumber)
	case 20:
		infoText = "Номер сваи: %s\n" +
			"Статус: запись в журнале;\n" +
			"Дата забивки: %s;\n" +
			"Факт. отметка верха головы: %d;\n" +
			"Оператор: %s;\n"
		infoText = fmt.Sprintf(infoText,
			p.PileNumber,
			p.StartDate.Format(time.DateOnly),
			p.FactPileHead,
			p.RecordedBy)
		if !p.CreatedAt.IsZero() {
			infoText = fmt.Sprintf("%sДата записи: %s;\n", infoText, p.CreatedAt.Format(time.DateTime))
		}
		if !p.UpdatedAt.IsZero() {
			infoText = fmt.Sprintf("%sДата изм.: %s\n", infoText, p.UpdatedAt.Format(time.DateTime))
		}
	}
	if title != "" {
		infoText = fmt.Sprintf("%s\n%s", title, infoText)
	}
	return infoText
}

func (b *TgBot) showPileInfo(chatID int64, title string) {
	b.sendMessage(chatID, b.getPileInfo(chatID, title))
}

func (b *TgBot) onBeforePileUpdateFPH(chatID int64) {
	p := b.userStates[chatID].currRec
	promt := ""
	if p.FactPileHead == 0 {
		promt = "Введи значение ФОВГ сваи (в мм, например, 10720):"
	} else {
		promt = fmt.Sprintf("Текущие значение ФОВГ %d мм. Введите новое значение (в мм):", p.FactPileHead)
	}
	b.sendMessage(chatID, promt)
	b.userStates[chatID].waitingFor = bm.WaitPileUpdateFPH
}

func (b *TgBot) onAfterPileUpdateFPH(chatID int64, data string) {
	num, err := strconv.Atoi(data)
	if err != nil {
		b.sendMessage(chatID, "Неверный формат ФОВГ. Пожалуйста, введите значение (в мм ):")
		return
	}
	b.userStates[chatID].currRec.FactPileHead = num
	b.insertOrUpdatePile(chatID)
}

func (b *TgBot) onAfterStartDateSelect(chatID int64, data string) {
	sd := time.Now()
	switch data {
	case bm.PileOpsStartDateYesterday:
		sd = time.Date(sd.Year(), sd.Month(), sd.Day()-1, 0, 0, 0, 0, time.UTC)
	case bm.PileOpsStartDateToday:
	default:
		panic("start date selection error: nor today neither yesterday")
	}
	b.userStates[chatID].currRec.StartDate = sd
	b.userStates[chatID].currRec.RecordedBy = b.userStates[chatID].userName
	b.insertOrUpdatePile(chatID)
}

func (b *TgBot) insertOrUpdatePile(chatID int64) {
	if b.userStates[chatID].currRec.StartDate.IsZero() {
		b.showPileStartDateMenu(chatID)
		b.userStates[chatID].waitingFor = bm.WaitPileStartDate
		return
	}
	if err := b.ws.InsertOrUpdatePdrLine(&b.userStates[chatID].currRec); err != nil {
		panic(err)
	}
	b.userStates[chatID].currRec.Status = 20
	b.showAfterUpdatePdrLineMenu(chatID)
}

func (b *TgBot) sendPdrLog(chatID int64, callback *tgbotapi.CallbackQuery) {
	if err := b.ws.SendPdrLog(b.userStates[chatID].currRec.ProjectId); err != nil {
		panic(err)
	}
	b.createAlert(callback, "Файл Excel отправлен на рабочий email")
}

func (b *TgBot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}
}

func (b *TgBot) createAlert(callback *tgbotapi.CallbackQuery, text string) {
	alert := tgbotapi.NewCallbackWithAlert(callback.ID, text)
	alert.ShowAlert = true
	if _, err := b.bot.Request(alert); err != nil {
		panic(fmt.Errorf("callback with alert error: %v", err))
	}
}
