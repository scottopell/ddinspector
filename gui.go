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
	statusObj               map[string]any
}

type IStatusApp struct {
	app                 *cview.Application
	df                  *DataFetcher
	state               *IStatusAppState
	dogstatsdUpdateChan chan DogstatsdPageProps
	overviewUpdateChan  chan OverviewPageProps
}

func (is *IStatusApp) InitState() {
	is.state = &IStatusAppState{
		dogstatsdCaptureEnabled: false,
		statusObj:               nil,
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

func (is *IStatusApp) SendOverviewProps() {
	props := OverviewPageProps{
		statusObj: is.df.statusJson(),
	}

	is.overviewUpdateChan <- props
}

func (is *IStatusApp) Run() {
	log.Print("Running application")
	refreshInterval, err := time.ParseDuration("1s")
	if err != nil {
		panic(err)
	}

	is.dogstatsdUpdateChan = make(chan DogstatsdPageProps)
	is.overviewUpdateChan = make(chan OverviewPageProps)
	overviewPage := NewOverviewPage(is.overviewUpdateChan, is.app.QueueUpdateDraw)
	dogstatsdPage := NewDogstatsdPage(is.dogstatsdUpdateChan, is.app.QueueUpdateDraw, is.SetDogstatsdCaptureEnabled)

	tabbedPanels := cview.NewTabbedPanels()

	go func() {
		go is.SendDogstatsdProps()
		for {
			is.state.statusObj = is.df.statusJson()
			go is.SendOverviewProps()
			time.Sleep(refreshInterval)
		}
	}()

	streamLogsPage := NewStreamLogsPage()

	tabbedPanels.AddTab("overview", "Overview", overviewPage.rootFlex)
	tabbedPanels.AddTab("dogstatsd", "Dogstatsd", dogstatsdPage.rootFlex)
	tabbedPanels.AddTab("stream_logs", "Stream Logs", streamLogsPage.rootFlex)

	root := tabbedPanels

	is.app.SetRoot(root, true)
	is.app.EnableMouse(true)
	is.app.SetFocus(root)
	if err := is.app.Run(); err != nil {
		panic(err)
	}
}
