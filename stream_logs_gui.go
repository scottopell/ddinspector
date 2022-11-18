package main

import (
	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"
)

type StreamLogsPage struct {
	rootFlex *cview.Flex
}

func NewStreamLogsPage() *StreamLogsPage {
	table := cview.NewTable()
	table.SetBorders(true)

	table.Select(0, 0)
	table.SetFixed(1, 0)

	tableCellSource := cview.NewTableCell("source")
	tableCellLog := cview.NewTableCell("source's log")
	tableCellSampleSource := cview.NewTableCell("sample app")
	tableCellSampleLog := cview.NewTableCell("03/22 08:51:06 TRACE  :...read_physical_netif: Home list entries returned = 7")

	table.SetCell(0, 0, tableCellSource)
	table.SetCell(0, 1, tableCellLog)
	table.SetCell(1, 0, tableCellSampleSource)
	table.SetCell(1, 1, tableCellSampleLog)

	table.SetSelectedFunc(func(row, column int) {
		table.GetCell(row, column).SetTextColor(tcell.ColorRed.TrueColor())
		table.SetSelectable(false, false)
	})

	parentFlex := cview.NewFlex()
	parentFlex.SetBorder(true)
	parentFlex.SetTitle("Streaming Logs")

	parentFlex.AddItem(table, 0, 1, true)

	streamLogs := StreamLogsPage{
		rootFlex: parentFlex,
	}

	return &streamLogs
}
