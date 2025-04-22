package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/excel/pkg/client"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/global"
	"github.com/obot-platform/tools/microsoft365/excel/pkg/graph"
)

func GetWorksheetColumnHeaders(ctx context.Context, workbookID, worksheetID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	data, _, err := graph.GetWorksheetColumnHeaders(ctx, c, workbookID, worksheetID)
	if err != nil {
		return err
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Println(string(dataBytes))
	return nil
}
