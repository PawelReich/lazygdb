package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PawelReich/gogdb/client"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/rivo/tview"
)

type SourceView struct {
	*CodeView
}

func NewSourceView(app *GoGdb) *SourceView {
	return &SourceView{CodeView: NewCodeView(app, "Source")}
}

func (view *SourceView) Update(frame *client.StoppedFrame) {

	fut := view.app.Debugger.GetCurrentStackFrame()
	view.app.Ui.QueueUpdateDraw(func() {
		stackFrame := <-fut
		if stackFrame.Error != nil {
			panic(stackFrame.Error)
		}

		view.SetTitle(stackFrame.Result.Frame.Function)

		view.Pane.SetText(view.PrettyPrintCode(&stackFrame.Result.Frame))

		lineInt, err := strconv.Atoi(stackFrame.Result.Frame.FileLine)
		if err != nil {
			view.app.LogError("Error parsing file line from stack frame: %s", stackFrame.Result.Frame.FileLine)
		}
		view.CenterView(lineInt)
	})
}

func (view *SourceView) PrettyPrintCode(frame *client.GdbStackFrame) string {
	var sb strings.Builder

	code, err := os.ReadFile(frame.FilePath)
	if err != nil {
		return fmt.Sprintf("[red::b]Could not read file: %s", frame.FilePath)
	}
	codeString := string(code)

	lexer := lexers.Match(frame.FilePath)

	iterator, err := lexer.Tokenise(nil, codeString)
	tokens := iterator.Tokens()
	lines := chroma.SplitTokensIntoLines(tokens)

	currentLine, err := strconv.Atoi(frame.FileLine)
	if err != nil {
		view.app.LogError("Error parsing file line from stack frame: %s", frame.FileLine)
		currentLine = -1
	}

	totalLines := len(strings.Split(codeString, "\n"))
	lineCounterWidth := len(strconv.Itoa(totalLines))

	for line, tokens := range lines {
		if currentLine == line {
			sb.WriteString("[white::ib]")
		} else {
			sb.WriteString("[grey::i]")
		}
		fmt.Fprintf(&sb, "%*d", lineCounterWidth, line)
		sb.WriteString(" [::I]│[::]")

		for _, token := range tokens {
			var colorTag string
			switch token.Type.Category() {
			case chroma.Keyword:
				colorTag = "[yellow]"
			case chroma.Name:
				colorTag = "[cyan]"
			case chroma.Literal:
				colorTag = "[green]"
			case chroma.Text:
				colorTag = "[magenta]"
			case chroma.Punctuation:
				colorTag = "[white]"
			case chroma.Comment:
				// // Hack: Constants (e.g. 0x80e8153c  add r2, r1, #304    @ 0x130)
				// // Are improperly tokenized as comments, force them as `Punctuation` for now
				// if token.Value[0] == '#' {
				// 	colorTag = "[white]"
				// } else {
				colorTag = "[grey]"
				// }

			default:
				colorTag = "[white]"
				// view.app.LogError(fmt.Sprintf("FAILED %s = %s\n", token.Type.Category(), tview.Escape(token.Value)))
			}
			sb.WriteString(colorTag)
			sb.WriteString(tview.Escape(token.Value))
		}
		sb.WriteString("[::B]")
	}

	return sb.String()
}
