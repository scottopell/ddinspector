package main

import (
	"fmt"

	"code.rocketnine.space/tslocum/cview"
)

type DogstatsdPage struct {
	rootFlex                       *cview.Flex
	displayTextView                *cview.TextView
	enableDogstatsdCaptureCheckBox *cview.CheckBox
}

type DogstatsdPageProps struct {
	dogstatsdCaptureEnabled bool
	dogstatsdData           map[string]any
}

func (dp *DogstatsdPage) update(data DogstatsdPageProps) {
	dp.enableDogstatsdCaptureCheckBox.SetChecked(data.dogstatsdCaptureEnabled)

	dp.displayTextView.Clear()
	if data.dogstatsdData != nil {
		fmt.Fprintf(dp.displayTextView, "metric | tags | value | timestamp")
		fmt.Fprintf(dp.displayTextView, "\n sys.cpu | fakehost | 0.4 | 2022-11-18....\n")
		fmt.Fprintf(dp.displayTextView, "\n sys.mem | fakehost | 3.6 | 2022-11-17....\n")
		fmt.Fprintf(dp.displayTextView, "\n sys.cpu | fakehost | 1.4 | 2022-11-18....\n")
		fmt.Fprintf(dp.displayTextView, "\n sys.cpu | fakehost | 0.5 | 2022-11-18....\n")
		fmt.Fprintf(dp.displayTextView, "\n sys.smth | fakehost | 1.6 | 2022-11-18....\n")
		fmt.Fprintf(dp.displayTextView, "\n sys.cpu | fakehost | 0.4 | 2022-11-18....\n")
	}
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
	parentFlex.AddItem(enableHeaderFlex, 0, 1, false)
	parentFlex.AddItem(mainFlex, 0, 4, false)

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
