package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/util"
)

func ListDriveItems(ctx context.Context, driveID string, itemID string) error {
	if driveID == "me" {
		fmt.Println("Error: drive_id must be the actual drive ID, cannot be 'me'")
		return nil
	}

	// Default to root folder if not specified
	if itemID == "" {
		itemID = "root" // "root" is a special identifier for the root folder
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	items, err := graph.ListDriveItems(ctx, c, driveID, itemID)
	if err != nil {
		return fmt.Errorf("failed to list drive items: %w", err)
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	var elements []gptscript.DatasetElement
	for _, item := range items {
		// Print item type indicator and name
		itemStr := ""
		if folder := item.GetFolder(); folder != nil {
			itemStr = "Item Type: 📁 folder "
		} else {
			itemStr = "Item Type: 📄 file "
		}
		itemStr += fmt.Sprintf("Name: %s\n", util.Deref(item.GetName()))

		// Print detailed information indented
		itemStr += fmt.Sprintf("ID: %s\n", util.Deref(item.GetId()))
		if size := item.GetSize(); size != nil {
			itemStr += fmt.Sprintf("Size: %d bytes\n", util.Deref(size))
		}
		if created := item.GetCreatedDateTime(); created != nil {
			itemStr += fmt.Sprintf("Created: %s\n", created.Format(time.RFC3339))
		}
		if modified := item.GetLastModifiedDateTime(); modified != nil {
			itemStr += fmt.Sprintf("Modified: %s\n", modified.Format(time.RFC3339))
		}
		if webUrl := item.GetWebUrl(); webUrl != nil {
			itemStr += fmt.Sprintf("Web URL: %s\n", util.Deref(webUrl))
		}

		// If it's a folder, show child count
		if folder := item.GetFolder(); folder != nil {
			itemStr += fmt.Sprintf("Child Count: %d\n", util.Deref(folder.GetChildCount()))
		}

		// If it's a file, show additional file properties
		if file := item.GetFile(); file != nil {
			if mimeType := file.GetMimeType(); mimeType != nil {
				itemStr += fmt.Sprintf("MIME Type: %s\n", util.Deref(mimeType))
			}
		}
		elements = append(elements, gptscript.DatasetElement{
			DatasetElementMeta: gptscript.DatasetElementMeta{
				Name:        util.Deref(item.GetId()),
				Description: util.Deref(item.GetName()),
			},
			Contents: itemStr,
		})
	}

	if len(elements) == 0 {
		fmt.Println("No drive items found")
		return nil
	}

	datasetID, err := gptscriptClient.CreateDatasetWithElements(ctx, elements, gptscript.DatasetOptions{
		Name:        fmt.Sprintf("%s_onedrive_drive_item", driveID),
		Description: "Drive items in drive " + driveID,
	})
	if err != nil {
		return fmt.Errorf("failed to create dataset with elements: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d drive items\n", datasetID, len(items))

	return nil
}
