package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/excel/pkg/client"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/global"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/graph"
)

func CreateWorksheet(ctx context.Context, workbookID, name string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return err
	}

	id, err := graph.CreateWorksheet(ctx, c, workbookID, name)
	if err != nil {
		return err
	}

	fmt.Printf("Worksheet created with ID: %s\n", id)
	return nil
}
