package tui

import (
	"sync"

	"github.com/PawelReich/lazygdb/client"
	"github.com/PawelReich/lazygdb/internal/db"
	"github.com/rivo/tview"
)

type LazyGdb struct {
	Ui    *tview.Application
	Pages *tview.Pages

	Db *db.Db

	CommandPrompt *CommandPrompt
	Debugger      *client.GdbClient

	Views []Updatable

	SourceView *SourceView

	EnableGdbNotificationLogging bool
}

type View struct {
	app *LazyGdb
}

func NewView(app *LazyGdb) *View {
	return &View{app: app}
}

func (app *LazyGdb) SetModal(modal tview.Primitive, width int, height int) {
	if modal == nil {
		app.Pages.RemovePage("modal")
		return
	}

	height += 2 // Account for borders
	width += 2  // Account for borders

	wrapper := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(modal, height, 1, true).
			AddItem(nil, 0, 1, false),
			width, 1, true).
		AddItem(nil, 0, 1, false)

	app.Pages.AddPage("modal", wrapper, true, true)
}

type Updatable interface {
	GetPane() tview.Primitive
	Update(*client.StoppedFrame)
}

func (app *LazyGdb) CycleFocus(direction int) {
	for i, view := range app.Views {
		if !view.GetPane().HasFocus() {
			continue
		}
		nextView := (i + direction) % len(app.Views)
		if nextView == -1 {
			nextView = len(app.Views) - 1
		}
		app.Ui.SetFocus(app.Views[nextView].GetPane())
		return
	}
}

func (app *LazyGdb) UpdateViews(frame *client.StoppedFrame) {
	var wg sync.WaitGroup

	for _, view := range app.Views {
		wg.Go(func() {
			view.Update(frame)
		})
	}

	wg.Wait()
	app.Ui.QueueUpdateDraw(func() {
	})
}
