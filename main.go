package main

import (
	"code.rocketnine.space/tslocum/cview"
)

func main() {
	df := NewDataFetcher("localhost", 5001)

	app := IStatusApp{
		app: cview.NewApplication(),
		df:  df,
	}

	app.InitState()
	app.Run()
}
