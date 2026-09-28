package tui

import (
	"os"
	"strings"

	"github.com/PawelReich/lazygdb/client"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"fmt"
)

const Prompt = "❯ "

type CommandPrompt struct {
	Pane *tview.Flex

	history *tview.TextView
	input   *tview.InputField

	app         *LazyGdb
	lastCommand string
}

func NewCommandPrompt(app *LazyGdb) *CommandPrompt {
	cmdHistory := tview.NewTextView()
	cmdHistory.SetDynamicColors(true)
	cmdHistory.SetScrollable(true)

	cmdPrompt := tview.NewInputField()
	cmdPrompt.SetLabel(Prompt)
	cmdPrompt.SetFieldBackgroundColor(tcell.ColorBlack)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	view := &CommandPrompt{app: app, history: cmdHistory, input: cmdPrompt, Pane: flex}

	cmdPrompt.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		command := cmdPrompt.GetText()
		if command == "" {
			command = view.lastCommand
		} else {
			view.lastCommand = command
		}

		fmt.Fprintf(cmdHistory, "[white]%s%s\n", Prompt, command)

		if command == "q" || command == "quit" {
			os.Exit(0)
		}

		if command[0] == '-' {
			// Drop '-' as it is assumed in `SendAsync`
			command = command[1:]
			splitCmd := strings.Split(command, " ")

			// Prepare MI command and its arguments
			command = splitCmd[0]
			splitCmd = splitCmd[1:]

			res := <-app.Debugger.SendAsync(command, splitCmd...)
			app.LogMap(res.Result)

		} else {
			res := <-app.Debugger.SendConsoleCommandAsync(command)
			app.LogInfo(res.Result)
		}
	})

	flex.SetBorder(true)
	flex.SetTitle("Command Prompt")
	flex.AddItem(cmdHistory, 0, 1, false)
	flex.AddItem(cmdPrompt, 1, 0, true)

	scrollLog := cmdHistory.InputHandler()
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight,
			tcell.KeyPgUp, tcell.KeyPgDn:
			scrollLog(event, func(tview.Primitive) {})
		}
		return event
	})

	return view
}

func (view *CommandPrompt) Update(_ *client.StoppedFrame) {}

func (view *CommandPrompt) GetPane() tview.Primitive {
	return view.Pane
}

func (view *CommandPrompt) LogColorf(color string, format string, args ...any) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}
	fmt.Fprintf(view.history, "[%s]%s[-]\n", color, message)

	view.input.SetText("")
	view.history.ScrollToEnd()
}
