package commands

import (
	"context"

	"github.com/obot-platform/tools/microsoft365/excel/pkg/client"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/global"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/printers"
)

func ListWorksheets(ctx context.Context, workbookID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	infos, err := graph.ListWorksheetsInWorkbook(ctx, c, workbookID)
	if err != nil {
		return err
	}

	printers.PrintWorksheetInfos(infos)
	return nil
}
