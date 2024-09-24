package view

import (
	"GoJob/db"
	"GoJob/web"
	"GoJob/xlog"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"time"
)

func setList(list *tview.List, stopLoading chan bool) {
	go func() {
		time.Sleep(time.Second * 3)

		sqlite := db.NewSqlite()
		data, err := sqlite.SelectData("jumpit", "")
		if err != nil {
			xlog.Logger.Error(err)
			stopLoading <- true
			return
		}

		for _, v := range data {
			name := v["name"].(string)
			description := v["description"].(string)
			list.AddItem(name, description, 0, nil)
		}

		stopLoading <- true
	}()
}

func setLoading(app *tview.Application, list *tview.List, index int, stopLoading chan bool) {
	go func() {
		loadingSymbols := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		loading := true
		for loading {
			for _, symbol := range loadingSymbols {
				select {
				case <-stopLoading:
					loading = false
					main, _ := list.GetItemText(index)
					list.SetItemText(index, main, "")
					app.Draw()
					return
				default:
					main, _ := list.GetItemText(index)
					list.SetItemText(index, main, fmt.Sprintf("Loading %s", symbol))
					app.Draw()
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
}

func Init() {
	app := tview.NewApplication()

	// layout
	header := tview.NewTextView().SetText("[#00B7EB]GO JOB[-]").SetTextAlign(tview.AlignCenter).SetDynamicColors(true)
	menuList := tview.NewList().
		AddItem("Crawl Data", "`Jumpit`에서 golang 공고를 새로 읽어옵니다.", 'c', nil).
		AddItem("Read Data", "읽어온 `Jumpit` 공고를 새로 읽어옵니다.", 'r', nil)
	jumpitList := tview.NewList()
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(menuList, 6, 0, true).
		AddItem(jumpitList, 0, 1, false)

	// Select event
	menuList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		switch index {
		case 0:
			web.CrawlJumpit()
			break
		case 1:
			stopLoading := make(chan bool)
			setLoading(app, menuList, 1, stopLoading)
			setList(jumpitList, stopLoading)
			break
		}
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'c':
			app.SetFocus(menuList)
		case 'z':
			app.SetFocus(jumpitList)
		}
		return event
	})

	if err := app.SetRoot(layout, true).Run(); err != nil {
		panic(err)
	}
}
