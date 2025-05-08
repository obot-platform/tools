package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
)

func DeletePermission(ctx context.Context, driveID string, itemID string, permissionID string) error {
	if driveID == "me" {
		return fmt.Errorf("drive_id must be the actual drive ID, cannot be 'me'")
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	err = graph.DeleteDriveItemPermission(ctx, c, driveID, itemID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to delete permissions: %w", err)
	}

	fmt.Printf("Successfully deleted permission for item %s in drive %s\n", itemID, driveID)
	return nil
}
