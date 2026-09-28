package tui

import (
	"bytes"
	"encoding/json"
	"sync"
)

var logMu sync.Mutex

func (app *GoGdb) LogColorf(color string, format string, args ...any) {
	logMu.Lock()
	app.CommandPrompt.LogColorf(color, format, args...)
	logMu.Unlock()
}

func (app *GoGdb) LogColor(color string, message string) {
	app.LogColorf(color, "%s", message)
}

func (app *GoGdb) LogInfo(message string) {
	app.LogColor("white", message)
}

func (app *GoGdb) LogInfof(format string, args ...any) {
	app.LogColorf("white", format, args...)
}

func (app *GoGdb) LogDebug(message string) {
	app.LogColor("grey", message)
}

func (app *GoGdb) LogDebugf(format string, args ...any) {
	app.LogColorf("grey", format, args...)
}

func (app *GoGdb) LogError(message string) {
	app.LogColor("red", message)
}

func (app *GoGdb) LogErrorf(format string, args ...any) {
	app.LogColorf("red", format, args...)
}

func (app *GoGdb) LogMap(value map[string]any) {
	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "\t")

	err := encoder.Encode(value)

	if err != nil {
		app.LogError(err.Error())
	}
	app.LogDebug(buf.String())
}
