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
