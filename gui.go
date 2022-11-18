package main

import (
	"log"
	"time"

	"code.rocketnine.space/tslocum/cview"
)

type IStatusApp struct {
	app *cview.Application
	df  *DataFetcher
}

func (is *IStatusApp) Run() {
	log.Print("Running application")
	refreshInterval, err := time.ParseDuration("1s")
	if err != nil {
		panic(err)
	}
	newStatusChan := make(chan map[string]any)
	overviewPage := NewOverviewPage(newStatusChan, is.app.QueueUpdateDraw)
	go func() {
		for {
			newStatusChan <- is.df.statusJson()
			time.Sleep(refreshInterval)
		}
	}()
	root := overviewPage.rootFlex
	is.app.SetRoot(root, true)
	is.app.EnableMouse(true)
	is.app.SetFocus(root)
	if err := is.app.Run(); err != nil {
		panic(err)
	}
}
