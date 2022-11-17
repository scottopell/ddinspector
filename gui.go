package main

import (
	"log"
	"time"

	"github.com/rivo/tview"
)

type IStatusApp struct {
	app *tview.Application
	df  *DataFetcher
}

func (is *IStatusApp) Run() {
	log.Print("Running application")
	refreshInterval, err := time.ParseDuration("1s")
	if err != nil {
		panic(err)
	}
	newStatusChan := make(chan map[string]any)
	overviewPage := NewOverviewPage(newStatusChan)
	go func() {
		for {
			is.app.QueueUpdateDraw(func() {
				newStatusChan <- is.df.statusJson()
			})
			time.Sleep(refreshInterval)
		}
	}()
	root := overviewPage.rootFlex
	if err := is.app.SetRoot(root, true).SetFocus(root).Run(); err != nil {
		panic(err)
	}
}
