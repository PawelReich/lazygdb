package tui

import (
	"sync"

	"github.com/PawelReich/lazygdb/client"
	"github.com/PawelReich/lazygdb/internal/db"
	"github.com/rivo/tview"
)

type LazyGdb struct {
	Ui *tview.Application
	Db *db.Db

	CommandPrompt *CommandPrompt
	Debugger      *client.GdbClient

	Views []Updatable

	SourceView *SourceView
}

type View struct {
	app *LazyGdb
}

func NewView(app *LazyGdb) *View {
	return &View{app: app}
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
	app.Ui.QueueUpdateDraw(func() {
		var wg sync.WaitGroup

		for _, view := range app.Views {
			wg.Go(func() {
				view.Update(frame)
			})
		}

		wg.Wait()
	})
}
