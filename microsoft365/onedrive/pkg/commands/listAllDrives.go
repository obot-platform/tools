package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/printer"
)

func ListAllDrives(ctx context.Context) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Get my personal drive
	drive, err := graph.GetMyDrive(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to get my drive: %w", err)
	}

	fmt.Println("My Drive:")
	printer.PrintDrive(drive, false)

	// Get all organization drives
	drives, err := graph.ListOrgDrives(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list drives: %w", err)
	}

	fmt.Println("\nAvailable Org Drives:")
	for _, drive := range drives {
		printer.PrintDrive(drive, false)
		fmt.Println("---")
	}

	return nil
}
