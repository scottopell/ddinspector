package gui

import (
	"log"
	"time"

	"github.com/scottopell/ddinspector/pkg/util"

	"code.rocketnine.space/tslocum/cview"
)

type QueueUpdateFunc func(func(), ...cview.Primitive)

type DDInspectorAppState struct {
	dogstatsdCaptureEnabled bool
	statusObj               map[string]any
	dogstatsdCaptureData    map[uint64]util.MetricStat
}

type DDInspectorApp struct {
	App                 *cview.Application
	Df                  *util.AgentDataFetcher
	state               *DDInspectorAppState
	updateInterval      time.Duration
	dogstatsdUpdateChan chan DogstatsdPageProps
	overviewUpdateChan  chan OverviewPageProps
	checksUpdateChan    chan ChecksPageProps
}

func (is *DDInspectorApp) InitState() {
	isDogstatsdCaptureEnabled := is.Df.GetDogstatsdCaptureEnablementValue()
	is.state = &DDInspectorAppState{
		dogstatsdCaptureEnabled: isDogstatsdCaptureEnabled,
		statusObj:               nil,
		dogstatsdCaptureData:    nil,
	}
}

func (is *DDInspectorApp) SetDogstatsdCaptureEnabled(enabled bool) {
	if err := is.Df.EnableDogstatsdCapture(); err != nil {
		// TODO display this error to the user in the UI
		log.Println("Unable to enable dogstatsd capture :(")
		return
	}
	is.state.dogstatsdCaptureEnabled = enabled
	// Send enablement to agent
	// Decide when to send props, should update loop do it?
	is.SendDogstatsdProps()
}

func (is *DDInspectorApp) SendDogstatsdProps() {
	props := DogstatsdPageProps{
		dogstatsdCaptureEnabled: is.state.dogstatsdCaptureEnabled,
	}
	if is.state.dogstatsdCaptureEnabled {
		props.dogstatsdData = is.state.dogstatsdCaptureData
	}
	is.dogstatsdUpdateChan <- props
}

func (is *DDInspectorApp) SendOverviewProps() {
	props := OverviewPageProps{
		statusObj: is.Df.StatusJson(),
	}

	is.overviewUpdateChan <- props
}

func (is *DDInspectorApp) SendChecksProps() {
	props := ChecksPageProps{
		statusObj: is.Df.StatusJson(),
	}

	is.checksUpdateChan <- props
}

func (is *DDInspectorApp) UpdateLoop() {
	go is.SendDogstatsdProps()
	go is.SendChecksProps()
	for {
		is.state.statusObj = is.Df.StatusJson()
		if is.state.dogstatsdCaptureEnabled {
			d, err := is.Df.FetchDogstatsdCaptureData()
			if err == nil {
				is.state.dogstatsdCaptureData = d
			} else {
				log.Println("Encountered error while reading dogstatsd capture data: ", err)
			}
		}
		go is.SendOverviewProps()
		go is.SendDogstatsdProps()
		go is.SendChecksProps()
		time.Sleep(is.updateInterval)
	}
}

func (is *DDInspectorApp) Run() {
	log.Print("Running application")
	refreshInterval, err := time.ParseDuration("1s")
	if err != nil {
		panic(err)
	}
	is.updateInterval = refreshInterval

	is.dogstatsdUpdateChan = make(chan DogstatsdPageProps)
	is.overviewUpdateChan = make(chan OverviewPageProps)
	is.checksUpdateChan = make(chan ChecksPageProps)
	overviewPage := NewOverviewPage(is.overviewUpdateChan, is.App.QueueUpdateDraw)
	dogstatsdPage := NewDogstatsdPage(is.dogstatsdUpdateChan, is.App.QueueUpdateDraw, is.SetDogstatsdCaptureEnabled)
	checksPage := NewChecksPage(is.checksUpdateChan, is.App.QueueUpdateDraw)
	streamLogsPage := NewStreamLogsPage()

	tabbedPanels := cview.NewTabbedPanels()

	tabbedPanels.AddTab("overview", "Overview", overviewPage.rootFlex)
	tabbedPanels.AddTab("checks", "Checks", checksPage.rootFlex)
	tabbedPanels.AddTab("dogstatsd", "Dogstatsd", dogstatsdPage.rootFlex)
	tabbedPanels.AddTab("stream_logs", "Stream Logs", streamLogsPage.rootFlex)

	root := tabbedPanels

	is.App.SetRoot(root, true)
	is.App.EnableMouse(true)
	is.App.SetFocus(root)
	go is.UpdateLoop()
	if err := is.App.Run(); err != nil {
		panic(err)
	}
}
