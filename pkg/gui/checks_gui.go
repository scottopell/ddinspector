package gui

import (
	"code.rocketnine.space/tslocum/cview"
	"github.com/DataDog/datadog-agent/pkg/status"
)

type ChecksPageProps struct {
	statusObj map[string]any
}

type ChecksPage struct {
	rootFlex              *cview.Flex
	checksListingTextView *cview.TextView
	newDataChan           chan ChecksPageProps
}

func (cp *ChecksPage) updateChecksListingTextView(stats map[string]any) {
	tv := cp.checksListingTextView
	tv.Clear()

	runnerStats := stats["runnerStats"]
	autoConfigStats := stats["autoConfigStats"]
	checkSchedulerStats := stats["checkSchedulerStats"]
	pyLoaderStats := stats["pyLoaderStats"]
	pythonInit := stats["pythonInit"]
	inventoriesStats := stats["inventories"]

	checkStats := make(map[string]interface{})
	checkStats["RunnerStats"] = runnerStats
	checkStats["pyLoaderStats"] = pyLoaderStats
	checkStats["pythonInit"] = pythonInit
	checkStats["AutoConfigStats"] = autoConfigStats
	checkStats["CheckSchedulerStats"] = checkSchedulerStats
	checkStats["OnlyCheck"] = ""
	checkStats["CheckMetadata"] = inventoriesStats

	ansiWriter := cview.ANSIWriter(tv)
	status.RenderStatusTemplate(ansiWriter, "/collector.tmpl", checkStats)
}

func (cp *ChecksPage) updatePanels(props ChecksPageProps) {
	statusObj := props.statusObj
	if statusObj == nil {
		return
	}
	cp.updateChecksListingTextView(statusObj)
}

func NewChecksPage(newDataChan chan ChecksPageProps, queueUpdateDraw func(func(), ...cview.Primitive)) *ChecksPage {
	checksListingTextView := cview.NewTextView()
	checksListingTextView.SetDynamicColors(true)

	middleFlex := cview.NewFlex()
	middleFlex.SetDirection(cview.FlexRow)
	middleFlex.AddItem(checksListingTextView, 0, 1, false)

	parentFlex := cview.NewFlex()
	parentFlex.AddItem(middleFlex, 0, 2, false)

	checks := ChecksPage{
		newDataChan:           newDataChan,
		checksListingTextView: checksListingTextView,
		rootFlex:              parentFlex,
	}

	go func() {
		for data := range newDataChan {
			queueUpdateDraw(func() {
				checks.updatePanels(data)
			})
		}
	}()

	return &checks
}
