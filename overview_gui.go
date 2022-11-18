package main

import (
	"embed"
	"fmt"
	"io"
	"path"
	"text/template"

	"code.rocketnine.space/tslocum/cview"
	"github.com/DataDog/datadog-agent/pkg/status"
)

type OverviewPageProps struct {
	statusObj map[string]any
}

type OverviewPage struct {
	rootFlex           *cview.Flex
	aggregatorTextView *cview.TextView
	dogstatsdTextView  *cview.TextView
	metadataTextView   *cview.TextView
	logsTextView       *cview.TextView
	newDataChan        chan OverviewPageProps
}

//go:embed templates
var templatesFS embed.FS

var fmap = status.Textfmap()

// Copy-Pasta from ./pkg/status/render.go
// TODO - re-use this function _and_ the templates
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

	renderStatusTemplate(tv, "/aggregator.tmpl", aggregator)
}

func (op *OverviewPage) updateDogstatsdTextView(statusObj map[string]any) {
	tv := op.dogstatsdTextView
	tv.Clear()
	aggregator := statusObj["dogstatsdStats"].(map[string]any)

	renderStatusTemplate(tv, "/dogstatsd.tmpl", aggregator)
}

func (op *OverviewPage) updateMetadataTextView(statusObj map[string]any) {
	tv := op.metadataTextView
	tv.Clear()

	title := fmt.Sprintf("Agent (v%s)", statusObj["version"])
	statusObj["title"] = title

	renderStatusTemplate(tv, "/header.tmpl", statusObj)
}

func (op *OverviewPage) updateLogTextView(statusObj map[string]any) {
	tv := op.logsTextView
	tv.Clear()
	logs := statusObj["logsStats"].(map[string]any)

	renderStatusTemplate(tv, "/logsagent.tmpl", logs)
}

func (op *OverviewPage) updatePanels(props OverviewPageProps) {
	statusObj := props.statusObj
	if statusObj == nil {
		return
	}
	op.updateAggregatorTextView(statusObj)
	op.updateDogstatsdTextView(statusObj)
	op.updateMetadataTextView(statusObj)
	op.updateLogTextView(statusObj)
}

func NewOverviewPage(newDataChan chan OverviewPageProps, queueUpdateDraw func(func(), ...cview.Primitive)) *OverviewPage {
	aggregatorTextView := cview.NewTextView()
	dogstatsdTextView := cview.NewTextView()
	metadataTextView := cview.NewTextView()
	logsTextView := cview.NewTextView()

	leftFlex := cview.NewFlex()
	leftFlex.SetDirection(cview.FlexRow)
	leftFlex.AddItem(aggregatorTextView, 0, 1, false)
	leftFlex.AddItem(dogstatsdTextView, 0, 1, false)

	middleFlex := cview.NewFlex()
	middleFlex.SetDirection(cview.FlexRow)
	middleFlex.AddItem(metadataTextView, 0, 1, false)

	rightFlex := cview.NewFlex()
	rightFlex.AddItem(logsTextView, 0, 1, false)

	parentFlex := cview.NewFlex()
	parentFlex.AddItem(leftFlex, 0, 1, false)
	parentFlex.AddItem(middleFlex, 0, 2, false)
	parentFlex.AddItem(rightFlex, 0, 1, false)

	overview := OverviewPage{
		newDataChan:        newDataChan,
		aggregatorTextView: aggregatorTextView,
		dogstatsdTextView:  dogstatsdTextView,
		metadataTextView:   metadataTextView,
		logsTextView:       logsTextView,
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
