package tui

import (
	"fmt"
	"strings"

	"github.com/PawelReich/lazygdb/client"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/rivo/tview"
)

type DisassemblyView struct {
	*CodeView
}

func NewDisassemblyView(app *LazyGdb) *DisassemblyView {
	return &DisassemblyView{CodeView: NewCodeView(app, "Disassembly")}
}

func (view *DisassemblyView) Update(frame *client.StoppedFrame) {
	fut := view.app.Debugger.DisassembleAroundPC(256)
	disas := <-fut
	if disas.Error != nil {
		view.SetError(disas.Error.Error())
	}

	prettyAssembly, pcLine, err := view.PrettyPrintDisassembly(&disas.Result, frame.Address)
	if err != nil {
		view.SetError(err.Error())
	}

	view.SetTitle(frame.Architecture)
	view.Pane.SetText(prettyAssembly)
	view.CenterView(pcLine)
}

func (view *DisassemblyView) PrettyPrintDisassembly(disas *client.GdbAsmDisassemblyPayload, pc string) (string, int, error) {
	var sb strings.Builder
	pcLine := 0

	lexer := lexers.Get("gas")

	for i, insn := range disas.AsmInsns {

		sym := <-view.app.Debugger.GetCachedSymbol(insn.Address)

		if insn.Address == pc {
			pcLine = i
			sb.WriteString("[white::ib]")
		} else {
			sb.WriteString("[grey::i]")

		}
		sb.WriteString(insn.Address)
		sb.WriteString(" [::I] ")
		fmt.Fprintf(&sb, "%*s", 25, sym.Result)
		sb.WriteString("│[::] ")

		iterator, err := lexer.Tokenise(nil, insn.Inst)
		if err != nil {
			return "", -1, err
		}

		for _, token := range iterator.Tokens() {
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
				// Hack: Constants (e.g. 0x80e8153c  add r2, r1, #304    @ 0x130)
				// Are improperly tokenized as comments, force them as `Punctuation` for now
				if token.Value[0] == '#' {
					colorTag = "[white]"
				} else {
					colorTag = "[grey]"
				}
			default:
				colorTag = "[white]"
				// view.app.LogError(fmt.Sprintf("FAILED %s = %s\n", token.Type.Category(), tview.Escape(token.Value)))
			}
			sb.WriteString(colorTag)
			sb.WriteString(tview.Escape(token.Value))
		}

		sb.WriteString("[::B]\n")
	}

	return sb.String(), pcLine, nil
}
