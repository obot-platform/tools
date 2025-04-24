package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/obot-platform/tools/microsoft365/excel/pkg/graph"
)

func AddWorksheetColumn(ctx context.Context, workbookID, worksheetID, columnID, contents string) error {
	if err := graph.AddWorksheetColumn(ctx, workbookID, worksheetID, columnID, strings.Split(contents, "|")); err != nil {
		return err
	}
	fmt.Println("Column added successfully")
	return nil
}
