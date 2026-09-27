package tui

import (
	"github.com/PawelReich/gogdb/client"
	"github.com/rivo/tview"
)

type GoGdb struct {
	Ui            *tview.Application
	CommandPrompt *CommandPrompt
	Debugger      *client.GdbClient

	Views []Updatable

	SourceView *SourceView
}

type View struct {
	app *GoGdb
}

func NewView(app *GoGdb) *View {
	return &View{app: app}
}

type Updatable interface {
	Update(*client.StoppedFrame)
}
