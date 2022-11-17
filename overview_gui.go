package main

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type OverviewPage struct {
	rootFlex           *tview.Flex
	aggregatorTextView *tview.TextView
	metadataTextView   *tview.TextView
	logsTextView       *tview.TextView
	newDataChan        chan map[string]any
}

func (op *OverviewPage) updateAggregatorTextView(statusObj map[string]any) {
	tv := op.aggregatorTextView
	tv.Clear()
	aggregator := statusObj["aggregatorStats"].(map[string]any)

	fmt.Fprintf(tv, "Checks Metric Samples: %.f\n", aggregator["ChecksMetricSample"].(float64))
	fmt.Fprintf(tv, "DogStatsD Metric Samples: %.f\n", aggregator["DogstatsdMetricSample"].(float64))
	fmt.Fprintf(tv, "DogStatsD Metric Contexts: %.f\n", aggregator["DogstatsdContexts"].(float64))
	fmt.Fprintf(tv, "Events: %.f\n", aggregator["Event"].(float64))
	fmt.Fprintf(tv, "Flushes: %.f\n", aggregator["NumberOfFlush"].(float64))
}

func (op *OverviewPage) updateMetadataTextView(statusObj map[string]any) {
	tv := op.metadataTextView
	tv.Clear()
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

}

func (op *OverviewPage) updatePanels(statusObj map[string]any) {
	op.updateAggregatorTextView(statusObj)
	op.updateMetadataTextView(statusObj)
}

func NewOverviewPage(newDataChan chan map[string]any) *OverviewPage {
	aggView := tview.NewTextView()
	metaView := tview.NewTextView()
	logsView := tview.NewTextView()

	flex := tview.NewFlex().
		AddItem(aggView, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(metaView, 0, 1, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)"), 0, 3, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom (5 rows)"), 5, 1, false), 0, 2, false).
		AddItem(logsView, 20, 1, false)

	overview := OverviewPage{
		newDataChan:        newDataChan,
		aggregatorTextView: aggView,
		metadataTextView:   metaView,
		logsTextView:       logsView,
		rootFlex:           flex,
	}

	go func() {
		for {
			data := <-newDataChan
			overview.updatePanels(data)
		}
	}()

	return &overview
}
