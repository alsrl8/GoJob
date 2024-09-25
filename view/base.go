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

func setList(list *tview.List) {
	sqlite := db.NewSqlite()
	data, err := sqlite.SelectData("jumpit", "")
	if err != nil {
		xlog.Logger.Error(err)
		return
	}

	for i, v := range data {
		name := v["name"].(string)
		description := v["description"].(string)
		list.AddItem(fmt.Sprintf("[#00B7EB]%d.[-] %s", i+1, name), description, 0, nil)
	}
}

func setLoadingForListView(app *tview.Application, list *tview.List, index int, stopLoading chan bool) {
	go func() {
		loadingSymbols := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		loading := true
		for loading {
			for _, symbol := range loadingSymbols {
				select {
				case <-stopLoading:
					loading = false
					main, _ := list.GetItemText(index)
					lastExecTime := time.Now().Format("2006-01-02 15:04:05")
					list.SetItemText(index, main, lastExecTime+" (Complete)")
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

func setLoadingForTextView(app *tview.Application, detail *tview.TextView, stopLoading chan bool) {
	go func() {
		loadingSymbols := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		loading := true
		for loading {
			for _, symbol := range loadingSymbols {
				select {
				case <-stopLoading:
					loading = false
					return
				default:
					detail.SetText(fmt.Sprintf("Loading %s", symbol))
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
		AddItem("Read Data", "읽어온 `Jumpit` 공고를 새로 불러옵니다.", 'r', nil)
	jumpitList := tview.NewList()
	jumpitDetail := tview.NewTextView()
	bodyLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(jumpitList, 0, 1, false).
		AddItem(jumpitDetail, 0, 1, false)
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(menuList, 6, 0, true).
		AddItem(bodyLayout, 0, 1, false)

	// Select event
	menuList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		switch index {
		case 0:
			stopLoading := make(chan bool)
			setLoadingForListView(app, menuList, 0, stopLoading)
			go func() {
				web.CrawlJumpit()
				stopLoading <- true
			}()
			break
		case 1:
			jumpitList.Clear()
			stopLoading := make(chan bool)
			setLoadingForListView(app, menuList, 1, stopLoading)
			go func() {
				time.Sleep(time.Second * 2)
				setList(jumpitList)
				stopLoading <- true
				app.SetFocus(jumpitList)
			}()
			break
		}
	})
	jumpitList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		stopLoading := make(chan bool)
		setLoadingForTextView(app, jumpitDetail, stopLoading)
		go func() {
			web.CrawlJumpitPostDetail(index, jumpitDetail, stopLoading)
			app.Draw()
		}()
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'c', 'r':
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
