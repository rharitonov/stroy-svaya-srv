package tgbot

import (
	"stroy-svaya/internal/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserState struct {
	UserId  int
	CurrRec model.PileDrivingRecordLine
}

type TgBot struct {
	token      string
	bot        *tgbotapi.BotAPI
	UserStates map[int64]*UserState
}

func NewTgBot(token string) *TgBot {
	return &TgBot{token: token}
}

type ICommandHandler interface {
	GetCommandName() string
	ShowCondition() bool
	GetInfo() string
	Run() error
}

type MenuBuilder struct {
	Menu map[string][]ICommandHandler
}

func NewMenuBuilder() *MenuBuilder {
	return &MenuBuilder{}
}

func (b *MenuBuilder) AddMenuCommand(parentCmdName string, cmdHandler ICommandHandler) {
	c := b.Menu[parentCmdName]
	c = append(c, cmdHandler)
	b.Menu[parentCmdName] = c
}

func (b *MenuBuilder) RunLoopMenu() {

}
