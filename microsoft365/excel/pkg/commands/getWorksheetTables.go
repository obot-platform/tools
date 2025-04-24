package commands

import (
	"context"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/client"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/global"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/printers"
)

func GetWorksheetTables(ctx context.Context, workbookID, worksheetID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	tables, err := graph.GetWorksheetTables(ctx, c, workbookID, worksheetID)
	if err != nil {
		return err
	}
	printers.PrintWorksheetTableInfos(tables)
	return nil
}
