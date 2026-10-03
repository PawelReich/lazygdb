package main

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/PawelReich/lazygdb/client"
	"github.com/PawelReich/lazygdb/internal/db"
	"github.com/PawelReich/lazygdb/tui"
	"github.com/spf13/pflag"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	gdb, err := client.New()
	if err != nil {
		panic(err)
	}

	ui := tview.NewApplication()
	db, err := db.NewDb()
	if err != nil {
		panic(err)
	}

	app := &tui.LazyGdb{Ui: ui, Debugger: gdb, Db: db}

	codeView := tui.NewSourceView(app)
	app.Views = append(app.Views, codeView)
	app.SourceView = codeView

	commandPrompt := tui.NewCommandPrompt(app)
	app.Views = append(app.Views, commandPrompt)
	app.CommandPrompt = commandPrompt

	diassemblyView := tui.NewDisassemblyView(app)
	app.Views = append(app.Views, diassemblyView)

	registersView := tui.NewRegistersView(app)
	app.Views = append(app.Views, registersView)

	stacktraceView := tui.NewStacktraceView(app)
	app.Views = append(app.Views, stacktraceView)

	expressionsView := tui.NewWatchView(app)
	app.Views = append(app.Views, expressionsView)
	app.ExpressionsView = expressionsView

	go func() {
		var errorLog string
		var consoleLog string

		for notification := range gdb.Notifications() {

			if notification["class"] == "stopped" {

				frame, err := gdb.ParseFrame(notification)
				if err != nil {
					panic(err)
				}
				app.UpdateViews(frame)
			}

			switch notification["type"] {
			case "target":
				fallthrough
			case "console":
				fallthrough
			case "log":
				consoleLog += notification["payload"].(string)
				for strings.Contains(consoleLog, "\n") {
					idx := strings.Index(consoleLog, "\n")
					log := consoleLog[:idx]
					consoleLog = consoleLog[idx+1:]
					slog.Info(log)
				}
			case "error":
				errorLog += notification["payload"].(string)
				for strings.Contains(errorLog, "\n") {
					idx := strings.Index(errorLog, "\n")
					log := errorLog[:idx]
					errorLog = errorLog[idx+1:]
					slog.Error(log)
				}
			}
			if app.EnableGdbNotificationLogging {
				slog.Info(fmt.Sprintf("%+v", notification))
			}
		}
	}()

	app.Ui.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			app.Ui.Stop()
		case tcell.KeyCtrlC:
			err := app.Debugger.Interrupt()
			if err != nil {
				slog.Error("Error while interrupting: " + err.Error())
			}
			return nil
		case tcell.KeyTab:
			app.CycleFocus(1)
			return nil
		case tcell.KeyBacktab:
			app.CycleFocus(-1)
			return nil
		}

		return event
	})

	grid := tview.NewGrid()
	grid.SetRows(0, 0, 20)
	grid.SetColumns(0, 0, 0, 0)

	// Code View (Top Left)
	// Starts at row 0, column 0. Spans 2 rows, 2 columns
	grid.AddItem(codeView.Pane, 0, 0, 2, 2, 0, 0, false)

	// Command Prompt (Bottom Left)
	// Starts at row 2, column 0. Spans 1 row, 2 columns.
	grid.AddItem(commandPrompt.Pane, 2, 0, 1, 2, 0, 0, true)

	// Disassembly (Top Right)
	// Starts at row 0, column 2. Spans 1 row, 2 columns.
	grid.AddItem(diassemblyView.Pane, 0, 2, 1, 2, 0, 0, false)

	// Registers (Middle Right)
	// Starts at row 1, column 2. Spans 1 row, 1 column.
	grid.AddItem(registersView.Pane, 1, 2, 1, 1, 0, 0, false)

	// Stacktrace (Bottom Mid-Right)
	// Starts at row 2, column 2. Spans 1 row, 2 columns.
	grid.AddItem(stacktraceView.Pane, 2, 2, 1, 2, 0, 0, false)

	// Expressions (Right Middle)
	// Starts at row 2, column 3. Spans 1 row, 1 column.
	grid.AddItem(expressionsView.Pane, 1, 3, 1, 1, 0, 0, false)

	var commands []string
	pflag.StringArrayVarP(&commands, "ex", "e", nil, "Commands to execute on launch")
	debugGdb := pflag.Bool("debuggdb", false, "Show GDB notifications")
	pflag.Parse()
	app.EnableGdbNotificationLogging = *debugGdb

	var commandWaitGroup sync.WaitGroup
	for _, command := range commands {
		commandWaitGroup.Go(func() {
			commandPrompt.HandleCommand(command)
		})
	}
	commandWaitGroup.Wait()

	app.Pages = tview.NewPages().
		AddPage("main", grid, true, true)

	err = app.Ui.SetRoot(app.Pages, true).SetFocus(commandPrompt.Pane).Run()
	if err != nil {
		panic(err)
	}
}
