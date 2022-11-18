package main

import (
	"fmt"
	"time"

	"code.rocketnine.space/tslocum/cview"
)

type DogstatsdPage struct {
	rootFlex                       *cview.Flex
	displayTextView                *cview.TextView
	enableDogstatsdCaptureCheckBox *cview.CheckBox
}

type DogstatsdPageProps struct {
	dogstatsdCaptureEnabled bool
	dogstatsdData           map[uint64]MetricStat
}

func (dp *DogstatsdPage) update(data DogstatsdPageProps) {
	dp.enableDogstatsdCaptureCheckBox.SetChecked(data.dogstatsdCaptureEnabled)

	dp.displayTextView.Clear()
	if data.dogstatsdData != nil {
		fmt.Fprintf(dp.displayTextView, "metric | tags | count | timestamp\n")
		for _, mstat := range data.dogstatsdData {
			fmt.Fprintf(dp.displayTextView, "| %s | %s | %d | %s|\n", mstat.Name, mstat.Tags, mstat.Count, mstat.LastSeen.Format(time.RFC3339))
		}
	}
}

func Separator() *cview.Box {
	b := cview.NewBox()
	b.SetBorder(true)
	return b
}

func NewDogstatsdPage(newPropsChan chan DogstatsdPageProps, queueUpdateDraw QueueUpdateFunc, setCaptureEnabled func(bool)) *DogstatsdPage {
	// Layout Initialization
	displayDogstatsdData := cview.NewTextView()
	enableDogstatsdCaptureText := cview.NewTextView()
	enableDogstatsdCaptureCheckBox := cview.NewCheckBox()
	fmt.Fprintf(enableDogstatsdCaptureText, "Enable DogStatsD Capture")

	enableDogstatsdCaptureCheckBox.SetChangedFunc(setCaptureEnabled)

	// Layout Configuration
	enableHeaderFlex := cview.NewFlex()
	enableHeaderFlex.SetDirection(cview.FlexColumn)
	enableHeaderFlex.AddItem(enableDogstatsdCaptureText, 0, 2, false)
	enableHeaderFlex.AddItem(enableDogstatsdCaptureCheckBox, 0, 1, false)
	enableHeaderFlex.AddItem(cview.NewBox(), 0, 2, false) // Filler

	mainFlex := cview.NewFlex()
	mainFlex.AddItem(displayDogstatsdData, 0, 1, false)

	parentFlex := cview.NewFlex()
	parentFlex.SetDirection(cview.FlexRow)
	parentFlex.AddItem(enableHeaderFlex, 1, 1, false)
	parentFlex.AddItem(Separator(), 1, 0, false)
	parentFlex.AddItem(mainFlex, 0, 2, false)

	dogstatsd := DogstatsdPage{
		displayTextView:                displayDogstatsdData,
		enableDogstatsdCaptureCheckBox: enableDogstatsdCaptureCheckBox,
		rootFlex:                       parentFlex,
	}

	go func() {
		for {
			data := <-newPropsChan
			queueUpdateDraw(func() {
				dogstatsd.update(data)
			})
		}
	}()

	return &dogstatsd
}
