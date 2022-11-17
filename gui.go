package main

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type IStatusApp struct {
	app *tview.Application
	df  *DataFetcher
}

func (is *IStatusApp) getStatusText() *tview.TextView {
	tv := tview.NewTextView()

	statusObj := is.df.statusJson()
	/*

		{
		  "ChecksHistogramBucketMetricSample": 0,
		  "ChecksMetricSample": 66355,
		  "ContainerLifecycleEvents": 0,
		  "ContainerLifecycleEventsErrors": 0,
		  "DogstatsdContexts": 0,
		  "DogstatsdContextsByMtype": {
			"Count": 0,
			"Counter": 0,
			"Distribution": 0,
			"Gauge": 0,
			"Histogram": 0,
			"Historate": 0,
			"MonotonicCount": 0,
			"Rate": 0,
			"Set": 0
		  },
		  "DogstatsdMetricSample": 1,
		  "Event": 1,
		  "EventPlatformEvents": {},
		  "EventPlatformEventsErrors": {},
		  "EventsFlushErrors": 0,
		  "EventsFlushed": 1,
		    },
		  "HostnameUpdate": 0,
		  "MetricTags": {
			"Series": {
			  "Above100": 0,
			  "Above90": 0
			},
			"Sketches": {
			  "Above100": 0,
			  "Above90": 0
			}
		  },
		  "NumberOfFlush": 452,
		  "OrchestratorManifests": 0,
		  "OrchestratorManifestsErrors": 0,
		  "OrchestratorMetadata": 0,
		  "OrchestratorMetadataErrors": 0,
		  "SeriesFlushErrors": 0,
		  "SeriesFlushed": 61211,
		  "ServiceCheck": 3186,
		  "ServiceCheckFlushErrors": 0,
		  "ServiceCheckFlushed": 3631,
		  "SketchesFlushErrors": 0,
		  "SketchesFlushed": 0
	*/
	aggregator := statusObj["aggregatorStats"].(map[string]any)

	fmt.Fprintf(tv, "Checks Metric Samples: %.f\n", aggregator["ChecksMetricSample"].(float64))
	fmt.Fprintf(tv, "DogStatsD Metric Samples: %.f\n", aggregator["DogstatsdMetricSample"].(float64))
	fmt.Fprintf(tv, "DogStatsD Metric Contexts: %.f\n", aggregator["DogstatsdContexts"].(float64))
	fmt.Fprintf(tv, "Events: %.f\n", aggregator["Event"].(float64))
	fmt.Fprintf(tv, "Flushes: %.f\n", aggregator["NumberOfFlush"].(float64))

	return tv
}

func (is *IStatusApp) getAgentMetadata() *tview.TextView {
	tv := tview.NewTextView()
	statusObj := is.df.statusJson()
	meta := statusObj["agent_metadata"].(map[string]any)
	hostinfo := statusObj["hostinfo"].(map[string]any)
	fmt.Fprintf(tv, "Version: %s\n", meta["agent_version"].(string))
	fmt.Fprintf(tv, "Flavor: %s\n", meta["flavor"].(string))
	fmt.Fprintf(tv, "Hostname: %s\n", hostinfo["hostname"].(string))

	if boottimeInt, ok := hostinfo["bootTime"].(float64); ok {
		boottime := time.Unix(int64(boottimeInt), 0)
		fmt.Fprintf(tv, "Boottime: %s", boottime.Format(time.RFC3339))
		if uptimeFloat, ok := hostinfo["uptime"].(float64); ok {
			fmt.Fprintf(tv, " (Been online %.f seconds)", uptimeFloat)
		}
		fmt.Fprintf(tv, "\n")
	}

	return tv
}

func (is *IStatusApp) getRoot() *tview.Flex {
	flex := tview.NewFlex().
		AddItem(is.getStatusText(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(is.getAgentMetadata(), 0, 1, false).
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
