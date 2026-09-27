package main

import (
	"fmt"
	"strings"

	"github.com/PawelReich/gogdb/client"
	"github.com/PawelReich/gogdb/tui"
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

	app := &tui.GoGdb{Ui: ui, Debugger: gdb}

	commandPrompt := tui.NewCommandPrompt(app)
	app.CommandPrompt = commandPrompt
	diassemblyView := tui.NewDisassemblyView(app)
	codeView := tui.NewSourceView(app)
	app.SourceView = codeView
	registersView := tui.NewRegistersView(app)
	stackView := tui.NewStackView(app)

	go func() {
		var errorLog string
		var consoleLog string

		for notification := range gdb.Notifications() {

			if notification["class"] == "stopped" {

				frame, err := gdb.ParseFrame(notification)
				if err != nil {
					panic(err)
				}

				go func() {
					diassemblyView.Update(frame)
				}()
				go func() {
					codeView.Update(frame)
				}()
				go func() {
					registersView.Update(frame)
				}()
				go func() {
					stackView.Update(frame)
				}()
			}

			switch notification["type"] {
			case "console":
				fallthrough
			case "log":
				consoleLog += notification["payload"].(string)
				for strings.Contains(consoleLog, "\n") {
					idx := strings.Index(consoleLog, "\n")
					log := consoleLog[:idx]
					consoleLog = consoleLog[idx+1:]
					app.LogInfo(log)
				}
			case "error":
				errorLog += notification["payload"].(string)
				for strings.Contains(errorLog, "\n") {
					idx := strings.Index(errorLog, "\n")
					log := errorLog[:idx]
					errorLog = errorLog[idx+1:]
					app.LogError(log)
				}
			}
			app.LogMap(notification)
		}
	}()

	app.Ui.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			app.Ui.Stop()
		case tcell.KeyCtrlC:
			app.Debugger.Interrupt()
			return nil
		case tcell.KeyTab:
			if diassemblyView.Pane.HasFocus() {
				app.Ui.SetFocus(codeView.Pane)
			} else if codeView.Pane.HasFocus() {
				app.Ui.SetFocus(commandPrompt.Pane)
			} else if commandPrompt.Pane.HasFocus() {
				app.Ui.SetFocus(stackView.Pane)
			} else if stackView.Pane.HasFocus() {
				app.Ui.SetFocus(diassemblyView.Pane)
			}
		}

		return event
	})

	grid := tview.NewGrid()
	// SetRows: Row 0 (dynamic), Row 1 (dynamic), Row 2 (fixed 3 lines for command prompt)
	grid.SetRows(0, 0, 20)

	// SetColumns: 3 columns of equal width (33% each)
	grid.SetColumns(0, 0, 0)

	// AddItem syntax:
	// AddItem(component, row, column, rowSpan, colSpan, minHeight, minWidth, focus)

	// 2. Code View (Top Left)
	// Spans 2 rows (0 and 1) and 2 columns (0 and 1)
	grid.AddItem(codeView.Pane, 0, 0, 2, 2, 0, 0, false)

	// 3. Command Prompt (Bottom Left)
	// Starts at row 2, spans 1 row tall, 2 columns wide
	grid.AddItem(commandPrompt.Pane, 2, 0, 1, 2, 0, 0, true)

	// 4. Disassembly (Top Right)
	// Starts at row 0, column 2. Spans 1 row, 1 column.
	grid.AddItem(diassemblyView.Pane, 0, 2, 1, 1, 0, 0, false)

	// 5. Registers (Middle Right)
	// Starts at row 1, column 2. Spans 1 row, 1 column.
	grid.AddItem(registersView.Pane, 1, 2, 1, 1, 0, 0, false)

	// 5. Registers (Middle Right)
	// Starts at row 2, column 2. Spans 1 row, 1 column.
	grid.AddItem(stackView.Pane, 2, 2, 1, 1, 0, 0, false)

	var commands []string
	pflag.StringArrayVarP(&commands, "ex", "e", nil, "Commands to execute")
	pflag.Parse()
	fmt.Println(commands)
	go func() {
		for _, command := range commands {
			ret := <-gdb.SendConsoleCommandAsync(command)

			if ret.Error != nil {
				app.LogError(err.Error())
			}

			app.LogInfo(ret.Result)
		}
	}()

	err = app.Ui.SetRoot(grid, true).SetFocus(commandPrompt.Pane).Run()
	if err != nil {
		panic(err)
	}

}
