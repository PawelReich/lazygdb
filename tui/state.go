package tui

import (
	"github.com/PawelReich/lazygdb/client"
	"github.com/rivo/tview"
)

type LazyGdb struct {
	Ui            *tview.Application
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

func (app *LazyGdb) CycleFocus() {
	for i, view := range app.Views {
		if !view.GetPane().HasFocus() {
			continue
		}
		nextView := (i + 1) % len(app.Views)
		app.Ui.SetFocus(app.Views[nextView].GetPane())
		return
	}
}
