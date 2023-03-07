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
	dogstatsdCaptureData    map[uint64]MetricStat
}

type IStatusApp struct {
	app                 *cview.Application
	df                  *AgentDataFetcher
	state               *IStatusAppState
	updateInterval      time.Duration
	dogstatsdUpdateChan chan DogstatsdPageProps
	overviewUpdateChan  chan OverviewPageProps
}

func (is *IStatusApp) InitState() {
	isDogstatsdCaptureEnabled := is.df.GetDogstatsdCaptureEnablementValue()
	is.state = &IStatusAppState{
		dogstatsdCaptureEnabled: isDogstatsdCaptureEnabled,
		statusObj:               nil,
		dogstatsdCaptureData:    nil,
	}
}

func (is *IStatusApp) SetDogstatsdCaptureEnabled(enabled bool) {
	if err := is.df.EnableDogstatsdCapture(); err != nil {
		// TODO display this error to the user in the UI
		log.Println("Unable to enable dogstatsd capture :(")
		return
	}
	is.state.dogstatsdCaptureEnabled = enabled
	// Send enablement to agent
	// Decide when to send props, should update loop do it?
	is.SendDogstatsdProps()
}

func (is *IStatusApp) SendDogstatsdProps() {
	props := DogstatsdPageProps{
		dogstatsdCaptureEnabled: is.state.dogstatsdCaptureEnabled,
	}
	if is.state.dogstatsdCaptureEnabled {
		props.dogstatsdData = is.state.dogstatsdCaptureData
	}
	is.dogstatsdUpdateChan <- props
}

func (is *IStatusApp) SendOverviewProps() {
	props := OverviewPageProps{
		statusObj: is.df.statusJson(),
	}

	is.overviewUpdateChan <- props
}

func (is *IStatusApp) UpdateLoop() {
	go is.SendDogstatsdProps()
	for {
		is.state.statusObj = is.df.statusJson()
		if is.state.dogstatsdCaptureEnabled {
			d, err := is.df.fetchDogstatsdCaptureData()
			if err == nil {
				is.state.dogstatsdCaptureData = d
			} else {
				log.Println("Encountered error while reading dogstatsd capture data: ", err)
			}
		}
		go is.SendOverviewProps()
		go is.SendDogstatsdProps()
		time.Sleep(is.updateInterval)
	}
}

func (is *IStatusApp) Run() {
	log.Print("Running application")
	refreshInterval, err := time.ParseDuration("1s")
	if err != nil {
		panic(err)
	}
	is.updateInterval = refreshInterval

	is.dogstatsdUpdateChan = make(chan DogstatsdPageProps)
	is.overviewUpdateChan = make(chan OverviewPageProps)
	overviewPage := NewOverviewPage(is.overviewUpdateChan, is.app.QueueUpdateDraw)
	dogstatsdPage := NewDogstatsdPage(is.dogstatsdUpdateChan, is.app.QueueUpdateDraw, is.SetDogstatsdCaptureEnabled)
	streamLogsPage := NewStreamLogsPage()

	tabbedPanels := cview.NewTabbedPanels()

	tabbedPanels.AddTab("overview", "Overview", overviewPage.rootFlex)
	tabbedPanels.AddTab("dogstatsd", "Dogstatsd", dogstatsdPage.rootFlex)
	tabbedPanels.AddTab("stream_logs", "Stream Logs", streamLogsPage.rootFlex)

	root := tabbedPanels

	is.app.SetRoot(root, true)
	is.app.EnableMouse(true)
	is.app.SetFocus(root)
	go is.UpdateLoop()
	if err := is.app.Run(); err != nil {
		panic(err)
	}
}
