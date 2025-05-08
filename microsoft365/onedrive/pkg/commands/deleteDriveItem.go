package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
)

func DeleteDriveItem(ctx context.Context, driveID string, itemID string) error {
	if driveID == "me" {
		fmt.Println("Error: drive_id must be the actual drive ID, cannot be 'me'")
		return nil
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := graph.DeleteDriveItem(ctx, c, driveID, itemID); err != nil {
		return fmt.Errorf("failed to delete drive item: %w", err)
	}

	fmt.Printf("Successfully deleted item %s from drive %s\n", itemID, driveID)
	return nil
}
