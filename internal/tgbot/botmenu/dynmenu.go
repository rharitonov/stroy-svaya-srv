package botmenu

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// type MenuState string

// const (
// 	PilesAll             MenuState = "PilesAll"
// 	PilesNew             MenuState = "PilesNew"
// 	PilesNoFPH           MenuState = "PilesNoFPH"
// 	PilesLoggedYesterday MenuState = "PilesLoggedYesterday"
// 	PilesLoggedToday     MenuState = "PilesLoggedToday"
// 	PilesSendExcel       MenuState = "PilesSendExcel"
// 	PileSelection        MenuState = "PileSelection"
// )

const (
	PilesAll             = "PilesAll"
	PilesNew             = "PilesNew"
	PilesNoFPH           = "PilesNoFPH"
	PilesLoggedYesterday = "PilesLoggedYesterday"
	PilesLoggedToday     = "PilesLoggedToday"
	PilesSendExcel       = "PilesSendExcel"
	PileSelection        = "PileSelection"
)

const (
	MaxMenuItems = 6
)

type DynamicMenu struct {
	CurrenMenu          map[string][]string
	AllElements         []string
	TgBotMenu           [][]tgbotapi.KeyboardButton
	TgBotMenuSingleItem string
}

func NewDynamicMenu(elements []string) *DynamicMenu {
	return &DynamicMenu{AllElements: elements}
}

func (dm *DynamicMenu) BuildMenuOrHandleSelection(param any) error {
	switch p := param.(type) {
	case []string:
		dm.buildMenu(p)
	case string:
		if strings.Contains(p, "..") {
			elements := dm.CurrenMenu[p]
			if elements == nil {
				return errors.New("неверный выбор группы. Попробуйте еще раз")
			}
			dm.buildMenu(elements)
			log.Printf("selected [..]")
		} else {
			if !slices.Contains(dm.AllElements, p) {
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

func (dm *DynamicMenu) buildMenu(elements []string) {
	dm.CurrenMenu = make(map[string][]string, MaxMenuItems)
	btnRows := make([][]tgbotapi.KeyboardButton, 0, MaxMenuItems)
	if len(elements) <= MaxMenuItems {
		for _, e := range elements {
			btn := tgbotapi.NewKeyboardButton(e)
			btnRows = append(btnRows, tgbotapi.NewKeyboardButtonRow(btn))
			se := make([]string, 1)
			se = append(se, e)
			dm.CurrenMenu[e] = se
		}
	} else {
		menuGroups := splitIntoMenuGroups(elements)
		for _, group := range menuGroups {
			if len(group) == 0 {
				continue
			}
			minElem := group[0]
			maxElem := group[len(group)-1]
			btnCaption := fmt.Sprintf("%s..%s", minElem, maxElem)
			btn := tgbotapi.NewKeyboardButton(btnCaption)
			btnRows = append(btnRows, tgbotapi.NewKeyboardButtonRow(btn))
			dm.CurrenMenu[btnCaption] = group
		}
	}
	dm.TgBotMenu = btnRows
	dm.TgBotMenuSingleItem = ""
}

func splitIntoMenuGroups(elements []string) [][]string {
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
