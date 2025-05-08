package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
)

func CreateFolder(ctx context.Context, driveID string, folderID string, folderName string) error {
	if driveID == "me" {
		fmt.Println("Error: drive_id must be the actual drive ID, cannot be 'me'")
		return nil
	}
	if folderName == "" {
		return fmt.Errorf("Error: folder_name is required but not provided")
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	item, err := graph.UploadDriveItem(ctx, c, driveID, folderID, folderName, nil, true)
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	// item, err := graph.CreateDriveFolder(ctx, c, driveID, folderID, folderName)
	// if err != nil {
	// 	return fmt.Errorf("failed to create folder: %w", err)
	// }

	fmt.Printf("Successfully created folder %s (ID: %s)\n", folderName, *item.GetId())
	if createdBy := item.GetCreatedBy().GetUser(); createdBy != nil {
		fmt.Printf("Folder Created by: %s\n", *createdBy.GetDisplayName())
	}
	if parentRef := item.GetParentReference(); parentRef != nil {
		fmt.Printf("Parent ID: %s\n", *parentRef.GetId())
		fmt.Printf("Parent Path: %s\n", *parentRef.GetPath())
	}
	if webUrl := item.GetWebUrl(); webUrl != nil {
		fmt.Printf("Web URL: %s\n", *webUrl)
	}
	return nil
}
