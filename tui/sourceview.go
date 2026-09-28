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

func (view *SourceView) Update(_ *client.StoppedFrame) {
	fut := view.app.Debugger.GetCurrentStackFrame()
	stackFramePayload := <-fut
	if stackFramePayload.Error != nil {
		view.app.LogErrorf("Error parsing stack frame payload: %s", stackFramePayload.Error)
		return
	}

	frame := stackFramePayload.Result.Frame

	if frame.FilePath == "" {
		view.SetError("No source available")
		return
	}

	fileLine, err := strconv.Atoi(frame.FileLine)
	if err != nil {
		view.SetError(fmt.Sprintf("Error parsing file line from stack frame: %s", frame.FileLine))
		return
	}
	view.RenderFile(frame.FilePath, fileLine, frame.Function)
}

func (view *SourceView) RenderFile(filePath string, fileLine int, function string) {

	prettyPrinted, err := view.PrettyPrintCode(filePath, fileLine)
	if err != nil {
		view.SetError(err.Error())
		return
	}
	view.SetTitle(function)
	view.Pane.SetText(prettyPrinted)
	view.CenterView(fileLine)
}

func (view *SourceView) PrettyPrintCode(filePath string, fileCurrentLine int) (string, error) {
	var sb strings.Builder

	code, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	codeString := string(code)

	lexer := lexers.Match(filePath)

	iterator, err := lexer.Tokenise(nil, codeString)

	if err != nil {
		return "", err
	}

	tokens := iterator.Tokens()
	lines := chroma.SplitTokensIntoLines(tokens)

	totalLines := len(strings.Split(codeString, "\n"))
	lineCounterWidth := len(strconv.Itoa(totalLines))

	for line, tokens := range lines {
		if fileCurrentLine == line+1 {
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

	return sb.String(), nil
}
