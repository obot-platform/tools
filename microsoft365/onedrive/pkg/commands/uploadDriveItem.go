package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/util"
)

func UploadDriveItem(ctx context.Context, driveID string, folderID string, workspaceFileName string) error {
	if driveID == "me" {
		fmt.Println("Error: drive_id must be the actual drive ID, cannot be 'me'")
		return nil
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	gs, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	content, err := gs.ReadFileInWorkspace(ctx, filepath.Join("files", workspaceFileName))
	if err != nil {
		return fmt.Errorf("failed to read file from workspace: %w", err)
	}

	item, err := graph.UploadDriveItem(ctx, c, driveID, folderID, workspaceFileName, content, false)
	if err != nil {
		return fmt.Errorf("failed to upload drive item: %w", err)
	}

	fmt.Printf("Successfully uploaded %s (ID: %s)\n", workspaceFileName, util.Deref(item.GetId()))
	if webUrl := item.GetWebUrl(); webUrl != nil {
		fmt.Printf("Web URL: %s\n", util.Deref(webUrl))
	}
	return nil
}
