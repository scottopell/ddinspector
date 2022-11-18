package main

import (
	"log"
	"time"

	"code.rocketnine.space/tslocum/cview"
)

type QueueUpdateFunc func(func(), ...cview.Primitive)

// Name ideas: inspector - cmd would be `agent inspector` which feels nice
type IStatusAppState struct {
	dogstatsdCaptureEnabled bool
}

type IStatusApp struct {
	app                 *cview.Application
	df                  *DataFetcher
	state               *IStatusAppState
	dogstatsdUpdateChan chan DogstatsdPageProps
}

func (is *IStatusApp) InitState() {
	is.state = &IStatusAppState{
		dogstatsdCaptureEnabled: false,
	}
}

func (is *IStatusApp) SetDogstatsdCaptureEnabled(enabled bool) {
	is.state.dogstatsdCaptureEnabled = enabled
	is.SendDogstatsdProps()
}

func (is *IStatusApp) SendDogstatsdProps() {
	props := DogstatsdPageProps{
		dogstatsdCaptureEnabled: is.state.dogstatsdCaptureEnabled,
	}
	if is.state.dogstatsdCaptureEnabled {
		props.dogstatsdData = map[string]any{
			"metricOne": true,
		}
	}
	is.dogstatsdUpdateChan <- props
}

func (is *IStatusApp) Run() {
	log.Print("Running application")
	refreshInterval, err := time.ParseDuration("1s")
	if err != nil {
		panic(err)
	}
	newStatusChan := make(chan map[string]any)
	is.dogstatsdUpdateChan = make(chan DogstatsdPageProps)
	_ = NewOverviewPage(newStatusChan, is.app.QueueUpdateDraw)

	dogstatsdPage := NewDogstatsdPage(is.dogstatsdUpdateChan, is.app.QueueUpdateDraw, is.SetDogstatsdCaptureEnabled)

	go func() {
		is.SendDogstatsdProps()
		for {
			newStatusChan <- is.df.statusJson()
			time.Sleep(refreshInterval)
		}
	}()

	root := dogstatsdPage.rootFlex
	is.app.SetRoot(root, true)
	is.app.EnableMouse(true)
	is.app.SetFocus(root)
	if err := is.app.Run(); err != nil {
		panic(err)
	}
}
