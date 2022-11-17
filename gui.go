package main

import (
	"fmt"

	"github.com/rivo/tview"
)

type IStatusApp struct {
	app *tview.Application
	df  *DataFetcher
}

func (is *IStatusApp) getStatusText() *tview.TextView {
	tv := tview.NewTextView().SetChangedFunc(func() {
		is.app.Draw()
	})

	statusObj := is.df.statusJson()
	fmt.Fprintf(tv, statusObj["version"].(string))

	return tv
}

func (is *IStatusApp) getRoot() *tview.Flex {
	flex := tview.NewFlex().
		AddItem(is.getStatusText(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Top"), 0, 1, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)"), 0, 3, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom (5 rows)"), 5, 1, false), 0, 2, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Right (20 cols)"), 20, 1, false)
	return flex
}

func (is *IStatusApp) Run() {
	root := is.getRoot()
	if err := is.app.SetRoot(root, true).SetFocus(root).Run(); err != nil {
		panic(err)
	}
}
