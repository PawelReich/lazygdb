package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PawelReich/lazygdb/client"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type StacktraceView struct {
	*View

	Pane *tview.List

	currentStack []client.GdbStackListFramesFrame
}

func NewStacktraceView(app *LazyGdb) *StacktraceView {

	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle("Stacktrace")
	list.ShowSecondaryText(true)
	list.SetWrapAround(false)
	list.SetSelectedBackgroundColor(tcell.ColorBlack)
	list.SetSelectedStyle(tcell.StyleDefault.Bold(true))

	view := &StacktraceView{View: NewView(app), Pane: list}

	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		frame := view.currentStack[index]

		fileLine, err := strconv.Atoi(frame.FileLine)
		if err != nil {
			app.SourceView.SetError(fmt.Sprintf("Error parsing file line from stack frame: %s", err.Error()))
			return
		}

		app.SourceView.RenderFile(frame.FilePath, fileLine, frame.Function)
	})

	return view
}

func (view *StacktraceView) Update(frame *client.StoppedFrame) {

	fut := view.app.Debugger.GetStacktrace()

	stacktrace := <-fut
	if stacktrace.Error != nil {
		view.SetError(stacktrace.Error.Error())
		return
	}

	view.currentStack = stacktrace.Result
	view.Pane.Clear()

	for _, stackFrame := range stacktrace.Result {
		function := stackFrame.Function
		if function == "" {
			function = "??"
		}

		stackEntry := fmt.Sprintf("[grey]> #%s [white]%s[grey::i]()", stackFrame.Level, tview.Escape(function))
		var stackSecondary strings.Builder
		for idx, arg := range stackFrame.Arguments {
			fmt.Fprintf(&stackSecondary, "%s=%s", arg.Name, arg.Value)
			if idx != len(stackFrame.Arguments)-1 {
				stackSecondary.WriteString(", ")
			}
		}

		view.Pane.AddItem(stackEntry, stackSecondary.String(), 0, nil)
	}
}

func (view *StacktraceView) GetPane() tview.Primitive {
	return view.Pane
}

func (view *StacktraceView) SetError(message string) {
	view.Pane.Clear()
	view.Pane.AddItem(fmt.Sprintf("[red]%s", tview.Escape(message)), "", 0, nil)
}
