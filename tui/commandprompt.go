package tui

import (
	"bytes"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/PawelReich/lazygdb/client"
	"github.com/PawelReich/lazygdb/internal"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const Prompt = "❯ "

type CommandPrompt struct {
	Pane *tview.Flex

	history       *tview.TextView
	historyBuffer bytes.Buffer
	historyMutex  sync.Mutex

	input *tview.InputField

	app                 *LazyGdb
	lastCommand         string
	historyScrollOffset int
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
		view.input.SetText("")
		view.HandleCommand(command)
	})

	flex.SetBorder(true)
	flex.SetTitle("Command Prompt")
	flex.AddItem(cmdHistory, 0, 1, false)
	flex.AddItem(cmdPrompt, 1, 0, true)

	scrollLog := cmdHistory.InputHandler()
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft, tcell.KeyRight,
			tcell.KeyPgUp, tcell.KeyPgDn:
			scrollLog(event, func(tview.Primitive) {})
		case tcell.KeyDown:
			view.ScrollHistory(-1)
		case tcell.KeyUp:
			view.ScrollHistory(1)
		default:
			view.historyScrollOffset = -1
		}
		return event
	})

	handler := internal.NewSimpleSlogHandler(view, slog.LevelDebug)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return view
}

func (view *CommandPrompt) Update(_ *client.StoppedFrame) {}

func (view *CommandPrompt) GetPane() tview.Primitive {
	return view.Pane
}

func (view *CommandPrompt) sendCommand(command string) {
	labelColor, _, _ := view.input.GetLabelStyle().Decompose()

	fmt.Fprintf(view, "[%s::b]%s[white::B]%s\n", labelColor, Prompt, command)

	if command[0] == '-' {
		// Drop '-' as it is assumed in `SendAsync`
		command = command[1:]
		splitCmd := strings.Split(command, " ")

		// Prepare MI command and its arguments
		command = splitCmd[0]
		splitCmd = splitCmd[1:]

		<-view.app.Debugger.SendAsync(command, splitCmd...)

	} else {
		res := <-view.app.Debugger.SendConsoleCommandAsync(command)
		slog.Info(res.Result)
	}
}

func (view *CommandPrompt) HandleCommand(command string) {
	if command == "" {
		if view.lastCommand == "" {
			return
		}
		command = view.lastCommand
	}
	switch command {
	case "":

	case "q", "quit":
		view.app.Ui.Stop()

	default:
		view.sendCommand(command)
	}
	view.lastCommand = command
	err := view.app.Db.InsertHistory(command)
	if err != nil {
		panic(err)
	}
}

func (view *CommandPrompt) ScrollHistory(direction int) {
	offset := view.historyScrollOffset + direction
	if offset < 0 {
		return
	}
	command, err := view.app.Db.GetCommandFromHistory(uint(offset))
	if err == sql.ErrNoRows {
		return
	}
	view.input.SetText(command)
	view.historyScrollOffset = offset
}

func (view *CommandPrompt) Write(p []byte) (int, error) {
	view.historyMutex.Lock()

	sz, err := view.historyBuffer.Write(p)

	view.historyMutex.Unlock()

	go view.app.Ui.QueueUpdateDraw(view.renderHistory)

	return sz, err
}

func (view *CommandPrompt) renderHistory() {
	view.historyMutex.Lock()

	view.history.Write(view.historyBuffer.Bytes())
	view.history.ScrollToEnd()
	view.historyBuffer.Reset()

	view.historyMutex.Unlock()
}
