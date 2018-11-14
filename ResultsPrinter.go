package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/table"
	"os"
)

func PrintResults() {
	fmt.Println()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	t.AppendHeader(table.Row{"Index Type", "Count"})
	t.AppendRows([]table.Row{
		{"Empty", emptyIndexCount},
		{"Future", futureIndexCount},
		{"Old", oldIndexCount},
	})
	t.AppendFooter(table.Row{"Total", emptyIndexCount + futureIndexCount + oldIndexCount})
	t.SetStyle(table.StyleColoredBlackOnBlueWhite)
	t.Render()
}
