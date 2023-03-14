package main

import (
	"github.com/scottopell/ddinspector/pkg/gui"
	"github.com/scottopell/ddinspector/pkg/util"

	"code.rocketnine.space/tslocum/cview"
)

func main() {
	df := util.NewAgentDataFetcher()
	if df == nil {
		return
	}
	app := gui.DDInspectorApp{
		App: cview.NewApplication(),
		Df:  df,
	}

	app.InitState()
	app.Run()
}
