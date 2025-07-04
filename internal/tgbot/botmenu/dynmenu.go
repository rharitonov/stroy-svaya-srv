package botmenu

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	PilesAll                  = "PilesAll"
	PilesNew                  = "PilesNew"
	PilesNoFPH                = "PilesNoFPH"
	PilesLoggedYesterday      = "PilesLoggedYesterday"
	PilesLoggedToday          = "PilesLoggedToday"
	PilesSendExcel            = "PilesSendExcel"
	PileOpsInsert             = "PileOpsInsert"
	PileOpsUpdateFPH          = "PileOpsUpdateFPH"
	PileOpsStartDateToday     = "PileOpsStartDateToday"
	PileOpsStartDateYesterday = "PileOpsStartDateYesterday"
	PileOpsBack               = "PileOpsBack"
	WaitPileNumber            = "WaitPileNumber"
	WaitPileOperation         = "WaitPileOperation"
	WaitPileUpdateFPH         = "WaitPileUpdateFPH"
	WaitPileStartDate         = "WaitPileStartDate"
)

const (
	MaxMenuItems        = 9
	MenuItemCountPerRow = 3
)

type DynamicMenu struct {
	CurrenMenu          map[string][]string
	allElements         []string
	TgBotMenuSingleItem string
}

func NewDynamicMenu(elements []string) *DynamicMenu {
	return &DynamicMenu{allElements: elements}
}

func (dm *DynamicMenu) BuildMenuOrHandleSelection(param any) error {
	switch p := param.(type) {
	case []string:
		dm.buildMenu(p)
	case nil:
		dm.buildMenu(dm.allElements)
	case string:
		if strings.Contains(p, "-") {
			elements := dm.CurrenMenu[p]
			if elements == nil {
				return errors.New("неверный выбор группы. Попробуйте еще раз")
			}
			dm.buildMenu(elements)
			log.Printf("selected [..]")
		} else {
			if !slices.Contains(dm.allElements, p) {
				return errors.New("неверный номер. Пожалуйста, выберите из предложенных вариантов")
			}
			dm.TgBotMenuSingleItem = p
		}
	default:
		return fmt.Errorf("unknown parameter %T: %p", p, p)
	}
	return nil
}

func (dm *DynamicMenu) SingleItemSelected() bool {
	return dm.TgBotMenuSingleItem != ""
}

func (dm *DynamicMenu) GetTgKeyboardMenu() tgbotapi.InlineKeyboardMarkup {
	kbRows := make([][]tgbotapi.InlineKeyboardButton, 0, MaxMenuItems/MenuItemCountPerRow)
	menu := dm.GetCurrentMenu()
	menuItemsCount := len(menu)
	if menuItemsCount == 0 {
		panic("no menu!")
	}
	btnRowsCount := menuItemsCount / MenuItemCountPerRow
	var btnPlusRow int = 0
	if (menuItemsCount % MenuItemCountPerRow) != 0 {
		btnPlusRow = 1
	}
	var s, e int = 0, 0
	for i := 0; i < btnRowsCount+btnPlusRow; i++ {
		if i == btnRowsCount {
			s = e
			e = menuItemsCount
		} else {
			s = e
			e = e + MenuItemCountPerRow
		}
		btns := make([]tgbotapi.InlineKeyboardButton, 0, MenuItemCountPerRow)
		for _, btnCaption := range menu[s:e] {
			btns = append(btns, tgbotapi.NewInlineKeyboardButtonData(btnCaption, btnCaption))
		}
		kbRows = append(kbRows, tgbotapi.NewInlineKeyboardRow(btns...))
	}
	btns := make([]tgbotapi.InlineKeyboardButton, 0, 1)
	btns = append(btns, tgbotapi.NewInlineKeyboardButtonData("В главное меню", PileOpsBack))
	kbRows = append(kbRows, tgbotapi.NewInlineKeyboardRow(btns...))
	return tgbotapi.NewInlineKeyboardMarkup(kbRows...)
}

func (dm *DynamicMenu) buildMenu(elements []string) {
	dm.CurrenMenu = make(map[string][]string, MaxMenuItems)
	if len(elements) <= MaxMenuItems {
		for _, e := range elements {
			se := make([]string, 1)
			se = append(se, e)
			dm.CurrenMenu[e] = se
		}
	} else {
		menuGroups := dm.splitIntoMenuGroups(elements)
		for _, group := range menuGroups {
			if len(group) == 0 {
				continue
			}
			minElem := group[0]
			maxElem := group[len(group)-1]
			btnCaption := ""
			if minElem == maxElem {
				btnCaption = minElem
			} else {
				btnCaption = fmt.Sprintf("%s-%s", minElem, maxElem)
			}
			dm.CurrenMenu[btnCaption] = group
		}
	}
	dm.TgBotMenuSingleItem = ""
}

func (dm *DynamicMenu) splitIntoMenuGroups(elements []string) [][]string {
	if len(elements) <= MaxMenuItems {
		result := make([][]string, len(elements))
		for i, e := range elements {
			result[i] = []string{e}
		}
		return result
	}
	groupSize := len(elements) / MaxMenuItems
	remain := len(elements) % MaxMenuItems
	groups := make([][]string, MaxMenuItems)
	start := 0
	for i := 0; i < MaxMenuItems; i++ {
		end := start + groupSize
		if i < remain {
			end++
		}
		groups[i] = elements[start:end]
		start = end
	}
	return groups
}

func (dm *DynamicMenu) GetCurrentMenu() []string {
	var result []string
	for caption := range dm.CurrenMenu {
		result = append(result, caption)
	}
	sort.Strings(result)
	return result
}
