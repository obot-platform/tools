package commands

import (
	"context"
	"fmt"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/printer"
)

func GetDrive(ctx context.Context, driveID string) error {
	if driveID == "me" {
		fmt.Println("Error: drive_id must be the actual drive ID, cannot be 'me'")
		return nil
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	var drive models.Driveable
	if driveID == "" {
		drive, err = graph.GetMyDrive(ctx, c)
		if err != nil {
			return fmt.Errorf("failed to get my personal drive: %w", err)
		}
	} else {
		drive, err = graph.GetDrive(ctx, c, driveID)
		if err != nil {
			return fmt.Errorf("failed to get drive: %w", err)
		}
	}
	printer.PrintDrive(drive, true)

	return nil
}
