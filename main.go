package main

import (
	"github.com/rivo/tview"
)

type IStatusConfig struct {
	statusUri     string
	authToken     string
	statusPort    int
	telemetryPort int
}

type IStatusApp struct {
	app    *tview.Application
	config *IStatusConfig
}

func (is *IStatusApp) getStatusText() *tview.TextView {

	// TODO
	return &tview.TextView{}
}

func (is *IStatusApp) getRoot() *tview.Flex {
	flex := tview.NewFlex().
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Left (1/2 x width of Top)"), 0, 1, false).
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

func main() {
	authToken := "TODO" // READ FROM DISK
	config := IStatusConfig{
		statusUri:     "localhost",
		statusPort:    5001,
		authToken:     authToken,
		telemetryPort: 5000,
	}

	app := IStatusApp{
		app:    tview.NewApplication(),
		config: &config,
	}
	app.Run()
}
