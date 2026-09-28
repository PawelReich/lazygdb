package tui

import (
	"fmt"

	"github.com/rivo/tview"
)

type CodeView struct {
	*View
	Pane *tview.TextView

	title string
}

func NewCodeView(app *GoGdb, title string) *CodeView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true)
	textView.SetScrollable(true)
	textView.SetBorder(true)

	view := &CodeView{View: NewView(app), title: title, Pane: textView}

	view.SetTitle("none")

	return view
}

func (view *CodeView) SetTitle(subtitle string) {
	view.Pane.SetTitle(tview.Escape(fmt.Sprintf("%s [%s]", view.title, subtitle)))
}

func (view *CodeView) CenterView(line int) {
	_, _, _, viewHeight := view.Pane.GetRect()
	middlePosition := line - viewHeight/2

	view.Pane.ScrollTo(middlePosition, 0)
}

func (view *CodeView) SetError(message string) {
	view.SetTitle("error")
	view.Pane.SetText(fmt.Sprintf("[red]%s", message))
}
