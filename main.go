package main

import (
	"code.rocketnine.space/tslocum/cview"
)

func main() {
	df := NewAgentDataFetcher()
	if df == nil {
		return
	}
	app := IStatusApp{
		app: cview.NewApplication(),
		df:  df,
	}

	app.InitState()
	app.Run()
}
