package slackbot

import (
	"fmt"

	ptable "github.com/jedib0t/go-pretty/table"
)

func Codeblock(message string) string {
	return fmt.Sprintf("```%s```", message)
}

func Table(headers []string, rows [][]string) string {
	t := ptable.NewWriter()
	tHeaders := make(ptable.Row, len(headers))

	for i, h := range headers {
		tHeaders[i] = h
	}

	t.AppendHeader(tHeaders)

	for _, r := range rows {
		tRow := make(ptable.Row, len(r))
		for i, v := range r {
			tRow[i] = v
		}
		t.AppendRow(tRow)
	}

	return t.Render()
}

func ResultSet(rows [][]interface{}) [][]string {
	var results = make([][]string, len(rows))

	for i, row := range rows {
		values := make([]string, len(row))
		for j, value := range row {
			values[j] = fmt.Sprint(value)
			results[i] = values
		}
	}

	return results
}
