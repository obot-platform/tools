package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/excel/pkg/client"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/global"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/printers"
)

func ListWorkbooks(ctx context.Context) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	workbookInfos, err := graph.ListWorkbooks(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list spreadsheets: %w", err)
	}

	printers.PrintWorkbookInfos(workbookInfos)
	return nil
}
