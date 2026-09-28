package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PawelReich/gogdb/client"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type StackView struct {
	*View

	Pane *tview.List

	currentStack []client.GdbStackListFramesFrame
}

func NewStackView(app *GoGdb) *StackView {

	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle("Stacktrace")
	list.ShowSecondaryText(true)
	list.SetWrapAround(false)
	list.SetSelectedBackgroundColor(tcell.ColorBlack)
	list.SetSelectedStyle(tcell.StyleDefault.Bold(true))

	view := &StackView{View: NewView(app), Pane: list}

	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		frame := view.currentStack[index]

		fileLine, err := strconv.Atoi(frame.FileLine)
		if err != nil {
			view.app.LogErrorf("Error parsing file line from stack frame: %s", frame.FileLine)
			fileLine = -1
		}

		app.SourceView.RenderFile(frame.FilePath, fileLine, frame.Function)
	})

	return view
}

func (view *StackView) Update(frame *client.StoppedFrame) {

	fut := view.app.Debugger.GetStacktrace()

	stacktrace := <-fut
	if stacktrace.Error != nil {
		view.app.LogErrorf("Error fetching stacktrace: %s", stacktrace.Error.Error())
		view.Pane.Clear()
		view.Pane.AddItem("[red::b]Error: "+tview.Escape(stacktrace.Error.Error()), "", 0, nil)
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
