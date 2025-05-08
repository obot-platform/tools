package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
)

func CopyDriveItem(ctx context.Context, sourceDriveID, sourceItemID, targetDriveID, targetFolderID, newName string) error {

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	err = graph.CopyDriveItem(ctx, c, sourceDriveID, sourceItemID, targetDriveID, targetFolderID, newName)
	if err != nil {
		return fmt.Errorf("failed to initiate copy operation: %w", err)
	}

	fmt.Printf("Successfully copied item to from %s/%s to %s/%s", sourceDriveID, sourceItemID, targetDriveID, targetFolderID)
	return nil
}
