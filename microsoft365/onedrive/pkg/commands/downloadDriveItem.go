package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
)

func DownloadDriveItem(ctx context.Context, driveID string, itemID string, workspaceFileName string) error {
	if driveID == "me" {
		fmt.Println("Error: drive_id must be the actual drive ID, cannot be 'me'")
		return nil
	}
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Get item details first to get the name
	item, err := graph.GetDriveItem(ctx, c, driveID, itemID)
	if err != nil {
		return fmt.Errorf("failed to get drive item details: %w", err)
	}

	content, err := graph.DownloadDriveItem(ctx, c, driveID, itemID)
	if err != nil {
		return fmt.Errorf("failed to download drive item: %w", err)
	}

	filename := workspaceFileName
	if filename == "" {
		filename = *item.GetName()
	}

	gs, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	if err := gs.WriteFileInWorkspace(ctx, filepath.Join("files", filename), content); err != nil {
		return fmt.Errorf("failed to write file to workspace: %w", err)
	}

	fmt.Printf("Successfully downloaded %s to workspace\n", filename)
	return nil
}
