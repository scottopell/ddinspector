package main

import (
	"embed"
	"fmt"
	"io"
	"path"
	"text/template"
	"time"

	"code.rocketnine.space/tslocum/cview"
	"github.com/DataDog/datadog-agent/pkg/status"
)

type OverviewPage struct {
	rootFlex           *cview.Flex
	aggregatorTextView *cview.TextView
	metadataTextView   *cview.TextView
	logsTextView       *cview.TextView
	newDataChan        chan map[string]any
}

//go:embed templates
var templatesFS embed.FS

var fmap = status.Textfmap()

// Copy-Pasta from ./pkg/status/render.go
func renderStatusTemplate(w io.Writer, templateName string, stats interface{}) {
	tmpl, tmplErr := templatesFS.ReadFile(path.Join("templates", templateName))
	if tmplErr != nil {
		fmt.Println(tmplErr)
		return
	}
	t := template.Must(template.New(templateName).Funcs(fmap).Parse(string(tmpl)))
	err := t.Execute(w, stats)
	if err != nil {
		fmt.Println(err)
	}
}

func (op *OverviewPage) updateAggregatorTextView(statusObj map[string]any) {
	tv := op.aggregatorTextView
	tv.Clear()
	aggregator := statusObj["aggregatorStats"].(map[string]any)

	// RENDER
	renderStatusTemplate(tv, "/aggregator.tmpl", aggregator)

	/*
		fmt.Fprintf(tv, "Checks Metric Samples: %.f\n", aggregator["ChecksMetricSample"].(float64))
		fmt.Fprintf(tv, "DogStatsD Metric Samples: %.f\n", aggregator["DogstatsdMetricSample"].(float64))
		fmt.Fprintf(tv, "DogStatsD Metric Contexts: %.f\n", aggregator["DogstatsdContexts"].(float64))
		fmt.Fprintf(tv, "Events: %.f\n", aggregator["Event"].(float64))
		fmt.Fprintf(tv, "Flushes: %.f\n", aggregator["NumberOfFlush"].(float64))
	*/
}

func (op *OverviewPage) updateMetadataTextView(statusObj map[string]any) {
	tv := op.metadataTextView
	tv.Clear()

	meta := statusObj["agent_metadata"].(map[string]any)
	hostinfo := statusObj["hostinfo"].(map[string]any)

	fmt.Fprintf(tv, "Host Metadata\n")
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

func (op *OverviewPage) updateLogTextView(statusObj map[string]any) {
	tv := op.logsTextView
	tv.Clear()
	logs := statusObj["logsStats"].(map[string]any)

	/*
		{
		  "is_running": true,
		  "endpoints": [
		    "Reliable: Sending compressed logs in HTTPS to agent-http-intake.logs.datadoghq.com on port 443"
		  ],
		  "metrics": {
		    "BytesSent": 0,
		    "EncodedBytesSent": 0,
		    "LogsProcessed": 0,
		    "LogsSent": 0
		  },
		  "integrations": [
		    {
		      "name": "logogo",
		      "sources": [
		        {
		          "bytes_read": 0,
		          "all_time_avg_latency": 0,
		          "all_time_peak_latency": 0,
		          "recent_avg_latency": 0,
		          "recent_peak_latency": 0,
		          "type": "file",
		          "configuration": {
		            "Path": "/tmp/logogo.log",
		            "Service": "logogo",
		            "Source": "dunnowhatthisisfor"
		          },
		          "status": "Error: cannot read file /tmp/logogo.log: stat /tmp/logogo.log: no such file or directory",
		          "inputs": [],
		          "messages": [],
		          "info": {}
		        }
		      ]
		    }
		  ],
		  "errors": [],
		  "warnings": [],
		  "use_http": true
		}


	*/
	fmt.Fprintf(tv, "Logs Agent\n")
	fmt.Fprintf(tv, "Running? %v", logs["is_running"].(bool))
}

func (op *OverviewPage) updatePanels(statusObj map[string]any) {
	op.updateAggregatorTextView(statusObj)
	op.updateMetadataTextView(statusObj)
	op.updateLogTextView(statusObj)
}

func NewOverviewPage(newDataChan chan map[string]any, queueUpdateDraw func(func(), ...cview.Primitive)) *OverviewPage {
	aggView := cview.NewTextView()
	metaView := cview.NewTextView()
	logsView := cview.NewTextView()

	placeholderBox := cview.NewBox()
	placeholderBox.SetBorder(true)
	placeholderBox.SetTitle("MiddleBlahBlah")

	parentFlex := cview.NewFlex()
	parentFlex.AddItem(aggView, 0, 1, false)
	middleFlex := cview.NewFlex()
	middleFlex.SetDirection(cview.FlexRow)
	middleFlex.AddItem(metaView, 0, 1, false)
	middleFlex.AddItem(placeholderBox, 0, 3, false)

	parentFlex.AddItem(middleFlex, 0, 2, false)
	parentFlex.AddItem(logsView, 0, 1, false)

	overview := OverviewPage{
		newDataChan:        newDataChan,
		aggregatorTextView: aggView,
		metadataTextView:   metaView,
		logsTextView:       logsView,
		rootFlex:           parentFlex,
	}

	go func() {
		for {
			data := <-newDataChan
			queueUpdateDraw(func() {
				overview.updatePanels(data)
			})
		}
	}()

	return &overview
}
