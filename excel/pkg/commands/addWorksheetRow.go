package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/tools/excel/pkg/client"
	"github.com/gptscript-ai/tools/excel/pkg/global"
	"github.com/gptscript-ai/tools/excel/pkg/graph"
)

func AddWorksheetRow(ctx context.Context, workbookID, worksheetID, contents string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return err
	}

	rows := strings.Split(contents, "\n")
	var parsedRows [][]string
	for _, row := range rows {
		parsedRows = append(parsedRows, strings.Split(row, "|"))
	}

	if err := graph.AddWorksheetRow(ctx, c, workbookID, worksheetID, parsedRows); err != nil {
		return err
	}

	fmt.Printf("Added %d rows successfully\n", len(parsedRows))
	return nil
}
