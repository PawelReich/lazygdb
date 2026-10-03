package tui

import (
	"fmt"

	"github.com/PawelReich/lazygdb/client"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type RegistersView struct {
	*View

	Pane         *tview.Table
	oldRegisters []client.Register
}

func NewRegistersView(app *LazyGdb) *RegistersView {
	textView := tview.NewTable()
	textView.SetBorder(true)
	textView.SetTitle("Registers")

	return &RegistersView{View: NewView(app), Pane: textView}
}

func (view *RegistersView) Update(_ *client.StoppedFrame) {
	registers := <-view.app.Debugger.GetRegisters()

	if registers.Error != nil {
		view.Pane.SetCellSimple(0, 0, "Error: "+registers.Error.Error())
		return
	}

	_, _, _, height := view.Pane.GetInnerRect()

	for i, reg := range registers.Result {
		row := i % height
		colGroup := i / height
		colOffset := colGroup * 2 // 2 columns per register (Name, Value)

		modifier := "::"
		if view.oldRegisters != nil && view.oldRegisters[i].Value != reg.Value {
			modifier = "red::b"
		}

		sym := <-view.app.Debugger.GetCachedSymbol(reg.Value)

		nameCell := tview.NewTableCell(fmt.Sprintf("[%s]%s[-]", modifier, reg.Name)).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignRight)

		valueCell := tview.NewTableCell(fmt.Sprintf("[%s]%s [grey::i]%s[-]", modifier, reg.Value, sym.Result)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft)

		view.Pane.SetCell(row, colOffset, nameCell)
		view.Pane.SetCell(row, colOffset+1, valueCell)

	}

	view.oldRegisters = registers.Result
}

func (view *RegistersView) GetPane() tview.Primitive {
	return view.Pane
}
