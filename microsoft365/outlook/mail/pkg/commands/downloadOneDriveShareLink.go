package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
)

func DownloadOneDriveShareLink(ctx context.Context, shareLink string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	filename, content, err := graph.GetOneDriveShareLink(ctx, c, shareLink)
	if err != nil {
		return fmt.Errorf("failed to get content of shared drive item: %w", err)
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
