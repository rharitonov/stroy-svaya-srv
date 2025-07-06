package tgbot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"stroy-svaya/internal/model"
	bm "stroy-svaya/internal/tgbot/botmenu"
	"stroy-svaya/internal/tgbot/webservice"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type pileRange struct {
	from model.PileDrivingRecordLine
	to   model.PileDrivingRecordLine
}

type UserState struct {
	projectID   int
	pileFieldID int
	user        model.User
	pdrRec      model.PileDrivingRecordLine
	pdrRange    pileRange
	menu        bm.DynamicMenu
	userName    string
	waitingFor  string
}

type TgBot struct {
	bot         *tgbotapi.BotAPI
	userStates  map[int64]*UserState
	ws          *webservice.WebService
	debugChatId int64
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
	b.setDebug(true)
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
			tgbotapi.NewInlineKeyboardButtonData("1️⃣ Выбрать сваю по номеру", bm.PileGetByNumber),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔢 Выбрать группу свай", bm.PileOpsInsertRange),
		),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("🔍 Все сваи", bm.PilesAll),
			tgbotapi.NewInlineKeyboardButtonData("Все сваи", bm.PilesAll),
			tgbotapi.NewInlineKeyboardButtonData("Незабитые", bm.PilesNew),
			tgbotapi.NewInlineKeyboardButtonData("Без ФОВГ", bm.PilesNoFPH),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Забитые сегодня", bm.PilesLoggedToday),
			tgbotapi.NewInlineKeyboardButtonData("Забитые вчера", bm.PilesLoggedYesterday),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Получить Excel", bm.PilesSendExcel),
		),
	)
	b.newInlineKb(chatID, &kb, "Добро пожаловать в журнал забивки свай!\n")
}

func (b *TgBot) showPileOperationsMenu(chatID int64) {
	baseRows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Ввод/изм. ФОВГ", bm.PileOpsUpdateFPH),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("В главное меню", bm.PileOpsBack),
		},
	}
	if b.userStates[chatID].pdrRec.Status == 10 {
		extraRow := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Запись в журнал", bm.PileOpsInsert),
		}
		baseRows = append([][]tgbotapi.InlineKeyboardButton{extraRow}, baseRows...)
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(baseRows...)
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
	b.newInlineKb(chatID, &kb, b.makePileInfoText(chatID, "Данные успешно записаны в журнал:"))
}

func (b *TgBot) startPileSelection(chatID int64, mode string) {
	piles, err := b.getPilesByFilter(chatID, mode)
	if err != nil {
		panic(err)
	}
	if len(piles) == 0 {
		b.sendMessage(chatID, "Отсутствуют сваи заданным критериям")
		return
	}
	b.userStates[chatID].menu = *bm.NewDynamicMenu(piles)
	b.makePileSelectionMenu(chatID, "")
	b.userStates[chatID].waitingFor = bm.WaitPileNumber
}

func (b *TgBot) getPilesByFilter(chatID int64, mode string) ([]string, error) {
	b.userStates[chatID].waitingFor = ""
	rec := model.PileDrivingRecordLine{}
	b.initPileWithCurrentPileField(chatID, &rec)
	rec.RecordedBy = b.userStates[chatID].userName

	b.userStates[chatID].pdrRec = rec
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
		return nil, err
	}
	return piles, nil
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
	case bm.WaitPileNumberInput:
		b.onAfterPileNumberInput(chatID, text)
	case bm.WaitPileNumberRange:
		b.onAfterPileNumberRangeInput(chatID, text)
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
	case bm.PileGetByNumber, bm.PileOpsInsertRange:
		b.onBeforePileNumberInput(chatID, data)
	default:
		switch b.userStates[chatID].waitingFor {
		case bm.WaitPileNumber:
			b.makePileSelectionMenu(chatID, data)
		case bm.WaitPileNumberRange:
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
		case bm.WaitPilesRangeStartDate:
			b.onAfterPilesRangeStartDateSelect(chatID, data)
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
	b.userStates[chatID].projectID = 1
	b.userStates[chatID].pileFieldID = 1
	if b.userStates[chatID].userName == "" {
		var err error
		var u *model.User
		u, err = b.ws.GetUserSetup(chatID)
		if err != nil {
			panic(err)
		}
		if u == nil {
			u = new(model.User)
			u.LastName = tgUser.LastName
			u.FirstName = tgUser.FirstName
			u.Initials = tgUser.FirstName
			u.TgUserId = chatID
			b.debugPrint(fmt.Sprintf("hello new user: %v", u))
		}
		b.userStates[chatID].user = *u
		b.userStates[chatID].userName = fmt.Sprintf("%s %s", u.LastName, u.Initials)
	}
}

func (b *TgBot) makePileSelectionMenu(chatID int64, data string) {
	if b.userStates[chatID].waitingFor == bm.WaitPileNumber {
		b.userStates[chatID].menu.BuildMenuOrHandleSelection(data)
	} else {
		b.userStates[chatID].menu.BuildMenuOrHandleSelection(nil)
	}
	if !b.userStates[chatID].menu.SingleItemSelected() {
		kb := b.userStates[chatID].menu.GetTgKeyboardMenu()
		b.newInlineKb(chatID, &kb, "Выберите номер сваи из предложенных вариантов:")
		return
	}
	b.updatePileRec(chatID, data)
	b.showPileInfo(chatID, "Выбрана свая:")
	b.showPileOperationsMenu(chatID)
	b.userStates[chatID].waitingFor = bm.WaitPileOperation
}

func (b *TgBot) initPileWithCurrentPileField(chatID int64, pile *model.PileDrivingRecordLine) {
	pile.ProjectId = b.userStates[chatID].projectID
	pile.PileFieldId = b.userStates[chatID].pileFieldID
}

func (b *TgBot) getPileRec(chatID int64, pile_no string) *model.PileDrivingRecordLine {
	filter := model.PileFilter{}
	filter.ProjectId = b.userStates[chatID].pdrRec.ProjectId
	filter.PileFieldId = b.userStates[chatID].pdrRec.PileFieldId
	filter.PileNumber = new(string)
	*filter.PileNumber = pile_no
	p, err := b.ws.GetPile(filter)
	if err != nil {
		panic(err)
	}
	return p
}

func (b *TgBot) updatePileRec(chatID int64, pile_no string) {
	b.userStates[chatID].pdrRec = *b.getPileRec(chatID, pile_no)
}

func (b *TgBot) makePileInfoText(chatID int64, title string) string {
	p := b.userStates[chatID].pdrRec
	infoText := ""
	switch p.Status {
	case 10:
		infoText = "Номер сваи: %s\nСтатус: нет записи в журнале"
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
	b.sendMessage(chatID, b.makePileInfoText(chatID, title))
}

func (b *TgBot) onBeforePileNumberInput(chatID int64, data string) {
	piles, err := b.getPilesByFilter(chatID, bm.PilesAll)
	if err != nil {
		panic(err)
	}
	if len(piles) == 0 {
		b.sendMessage(chatID, "Отсутствуют сваи с заданными критериями")
		return
	}
	b.userStates[chatID].menu = *bm.NewDynamicMenu(piles)
	switch data {
	case bm.PileGetByNumber:
		b.userStates[chatID].waitingFor = bm.WaitPileNumberInput
		b.sendMessage(chatID, "Введите номер сваи:")
	case bm.PileOpsInsertRange:
		b.userStates[chatID].waitingFor = bm.WaitPileNumberRange
		b.sendMessage(chatID, "Введите номер первой и последней сваи в группе через пробел:")
	}
}

func (b *TgBot) onAfterPileNumberInput(chatID int64, text string) {
	if !b.userStates[chatID].menu.Contains(text) {
		b.sendMessage(chatID, "Введенный номер сваи отсутствует в свайном поле. Введите номер еще раз:")
		return
	}
	b.updatePileRec(chatID, text)
	b.showPileInfo(chatID, "Выбрана свая:")
	b.showPileOperationsMenu(chatID)
	b.userStates[chatID].waitingFor = bm.WaitPileOperation
}

func (b *TgBot) onAfterPileNumberRangeInput(chatID int64, text string) {
	var failed bool
	r := pileRange{}
	piles := strings.Fields(text)
	if len(piles) != 2 {
		b.sendMessage(chatID, "Неверный формат. Введите номер первой и последней сваи в группе через пробел:")
		return
	}
	r.from.PileNumber = piles[0]
	r.to.PileNumber = piles[1]
	if !b.userStates[chatID].menu.Contains(r.from.PileNumber) {
		b.sendMessage(chatID, fmt.Sprintf("Номер сваи %s отсутствует в свайном поле. Введите группу еще раз:", r.from.PileNumber))
		failed = true
	}
	if !b.userStates[chatID].menu.Contains(r.to.PileNumber) {
		b.sendMessage(chatID, fmt.Sprintf("Номер сваи %s отсутствует в свайном поле. Введите группу еще раз:", r.to.PileNumber))
		failed = true
	}
	if failed {
		return
	}

	if r.from.PileNumber == r.to.PileNumber {
		b.sendMessage(chatID, "Неверный формат. Введите номер первой и последней сваи в группе через пробел:")
		return
	}

	b.userStates[chatID].pdrRange = r
	if !b.validatePileNumberRange(chatID) {
		return
	}

	b.showPileStartDateMenu(chatID)
	b.userStates[chatID].waitingFor = bm.WaitPilesRangeStartDate
}

func (b *TgBot) validatePileNumberRange(chatID int64) bool {
	piles, err := b.userStates[chatID].menu.GetRange(
		b.userStates[chatID].pdrRange.from.PileNumber,
		b.userStates[chatID].pdrRange.to.PileNumber)
	if err != nil {
		panic(err)
	}
	var loggedPiles []string
	for _, n := range piles {
		p := b.getPileRec(chatID, n)
		if p.Status != 10 {
			loggedPiles = append(loggedPiles, p.PileNumber)
		}
	}
	ln := len(loggedPiles)
	if ln != 0 {
		if ln == 1 {
			b.sendMessage(chatID, fmt.Sprintf("Вы выбраной группе имеется забитая свая с номером %s "+
				"Введите группу, исключающаю данную сваю:", loggedPiles[0]))
		} else {
			b.sendMessage(chatID, fmt.Sprintf("Вы выбраной группе имеются забитые сваи в количестве %d шт, "+
				"первый и посл. номера которых %s и %s соответственно. "+
				"Введите группу, исключающаю данные сваи:", ln, loggedPiles[0], loggedPiles[ln-1]))
		}
		return false
	}
	return true
}

func (b *TgBot) onBeforePileUpdateFPH(chatID int64) {
	p := b.userStates[chatID].pdrRec
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
	b.userStates[chatID].pdrRec.FactPileHead = num
	b.insertOrUpdatePile(chatID)
}

func (b *TgBot) onAfterStartDateSelect(chatID int64, data string) {
	b.userStates[chatID].pdrRec.StartDate = b.makeStartDate(data)
	b.userStates[chatID].pdrRec.RecordedBy = b.userStates[chatID].userName
	b.insertOrUpdatePile(chatID)
	b.userStates[chatID].waitingFor = ""
}

func (b *TgBot) onAfterPilesRangeStartDateSelect(chatID int64, data string) {
	if !b.validatePileNumberRange(chatID) {
		return
	}
	sd := b.makeStartDate(data)
	pilesNo, err := b.userStates[chatID].menu.GetRange(
		b.userStates[chatID].pdrRange.from.PileNumber,
		b.userStates[chatID].pdrRange.to.PileNumber)
	if err != nil {
		panic(err)
	}
	var piles []model.PileDrivingRecordLine
	for _, n := range pilesNo {
		p := b.getPileRec(chatID, n)
		p.StartDate = sd
		p.RecordedBy = b.userStates[chatID].userName
		if p.Status == 10 {
			piles = append(piles, *p)
		}
	}
	ln := len(piles)
	if ln == 0 {
		panic("no piles range to insert")
	}
	for _, p := range piles {
		if err := b.ws.InsertOrUpdatePdrLine(&p); err != nil {
			panic(err)
		}
	}
	b.sendMessage(chatID, fmt.Sprintf("Данные по %d сваям были успешно записаны в журнал", len(piles)))
	b.userStates[chatID].waitingFor = ""
	b.debugPrint(fmt.Sprintf("range inserted. chatID: %d; %s", chatID, b.userStates[chatID].userName))
}

func (b *TgBot) makeStartDate(data string) time.Time {
	sd := time.Now()
	switch data {
	case bm.PileOpsStartDateYesterday:
		sd = time.Date(sd.Year(), sd.Month(), sd.Day()-1, 0, 0, 0, 0, time.UTC)
	case bm.PileOpsStartDateToday:
	default:
		panic("start date selection error: nor today neither yesterday")
	}
	return sd
}

func (b *TgBot) insertOrUpdatePile(chatID int64) {
	if b.userStates[chatID].pdrRec.StartDate.IsZero() {
		b.showPileStartDateMenu(chatID)
		b.userStates[chatID].waitingFor = bm.WaitPileStartDate
		return
	}
	if err := b.ws.InsertOrUpdatePdrLine(&b.userStates[chatID].pdrRec); err != nil {
		panic(err)
	}
	b.userStates[chatID].pdrRec.Status = 20
	b.showAfterUpdatePdrLineMenu(chatID)
	b.debugPrint(fmt.Sprintf("pile inserted/updated. chatID: %d; %v", chatID, b.userStates[chatID].pdrRec))
}

func (b *TgBot) sendPdrLog(chatID int64, callback *tgbotapi.CallbackQuery) {
	b.debugPrint(fmt.Sprintf("sending excel. chatID: %d; %s", chatID, b.userStates[chatID].userName))
	if b.userStates[chatID].user.Email == "" {
		b.createAlert(callback, "Отправка файла Excel журнала возможна только для зарегестрированных пользователей. Обратитесь к администратору.")
		return
	}
	if err := b.ws.SendPdrLog(b.userStates[chatID].pdrRec.ProjectId, chatID); err != nil {
		panic(err)
	}
	b.createAlert(callback, "Файл Excel отправлен на рабочий email")
	b.userStates[chatID].waitingFor = ""
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

func (b *TgBot) debugPrint(text string) {
	if b.debugChatId == 0 {
		return
	}
	b.sendMessage(b.debugChatId, fmt.Sprintf("%s>> %s", time.Now().Format(time.DateTime), text))
}

func (b *TgBot) setDebug(debugMode bool) {
	b.bot.Debug = debugMode
	if debugMode {
		text := os.Getenv("DEBUGCHATID")
		chatId, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			log.Println(err)
			return
		}
		b.debugChatId = chatId
	}
}
