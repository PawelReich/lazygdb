package tui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/PawelReich/lazygdb/client"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ExpressionsView struct {
	*View

	Pane *tview.List

	expressions       []string
	expressionResults []string
}

func NewWatchView(app *LazyGdb) *ExpressionsView {
	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle("Expressions")
	list.ShowSecondaryText(false)

	focusedStyle := tcell.StyleDefault.Bold(true).Background(tcell.ColorDefault).Foreground(tcell.ColorWhite)
	blurredStyle := tcell.StyleDefault.Bold(true).Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)

	list.SetSelectedStyle(blurredStyle)
	list.SetFocusFunc(func() {
		list.SetSelectedStyle(focusedStyle)
	})
	list.SetBlurFunc(func() {
		list.SetSelectedStyle(blurredStyle)
	})

	view := &ExpressionsView{View: NewView(app), Pane: list}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'n' {
			view.app.SetModal(view.addExpressionModal(), 20, 1)
			return nil
		}
		return event
	})

	return view
}

func (view *ExpressionsView) Update(_ *client.StoppedFrame) {
	view.updateExpressions()
}

func (view *ExpressionsView) updateExpressions() {
	var wg sync.WaitGroup

	results := make([]string, len(view.expressions))

	for i, expr := range view.expressions {
		wg.Go(func() {
			results[i] = view.evaluateExpression(expr)
		})
	}

	wg.Wait()
	oldResults := view.expressionResults
	view.expressionResults = results

	view.renderExpressions(oldResults)
	view.expressionResults = results
}

func (view *ExpressionsView) evaluateExpression(expr string) string {
	res := <-view.app.Debugger.SendConsoleCommandAsync(expr)

	var result string
	if res.Error != nil {
		result = res.Error.Error()
	} else {
		if res.Result != "" {
			result = res.Result[strings.Index(res.Result, " = ")+3:]
		} else {
			result = "Error while evaluating"
		}
	}
	return result
}

func (view *ExpressionsView) renderExpressions(oldResults []string) {
	view.Pane.Clear()

	longestExpression := 0
	for _, expr := range view.expressions {
		if len(expr) > longestExpression {
			longestExpression = len(expr)
		}
	}

	for i, expr := range view.expressions {
		expr = tview.Escape(expr)
		result := tview.Escape(view.expressionResults[i])

		modifier := "yellow::"
		if oldResults != nil && oldResults[i] != result {
			modifier = "red::b"
		}

		formattedResult := fmt.Sprintf("[%s]%-*s[-] │ %s", modifier, longestExpression, expr, result)

		view.Pane.AddItem(formattedResult, "", 0, nil)
	}
}

func (view *ExpressionsView) GetPane() tview.Primitive {
	return view.Pane
}

func (view *ExpressionsView) addExpressionModal() tview.Primitive {
	modal := tview.NewFlex()
	modal.SetTitle("Enter new expression")
	modal.SetBorder(true)
	modal.SetDirection(tview.FlexRow)

	input := tview.NewInputField()
	input.SetFieldBackgroundColor(tcell.ColorBlack)
	input.SetDoneFunc(func(key tcell.Key) {
		view.app.SetModal(nil, 0, 0)

		view.AddExpression(input.GetText())
	})
	modal.AddItem(input, 0, 1, true)
	return modal
}

func (view *ExpressionsView) AddExpression(expr string) {
	view.expressions = append(view.expressions, expr)
	view.expressionResults = append(view.expressionResults, view.evaluateExpression(expr))

	view.renderExpressions(nil)
}

