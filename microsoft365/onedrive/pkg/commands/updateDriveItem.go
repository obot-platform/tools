package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/util"
)

func UpdateDriveItem(ctx context.Context, driveID string, itemID string, newFolderID string, newName string) error {
	if driveID == "me" {
		fmt.Println("Error: drive_id must be the actual drive ID, cannot be 'me'")
		return nil
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	item, err := graph.UpdateDriveItem(ctx, c, driveID, itemID, newFolderID, newName)
	if err != nil {
		return fmt.Errorf("failed to get drive item: %w", err)
	}

	fmt.Printf("Successfully updated drive item %s (ID: %s)\n", util.Deref(item.GetName()), util.Deref(item.GetId()))

	return nil
}
