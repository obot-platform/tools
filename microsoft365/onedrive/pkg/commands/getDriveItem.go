package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/util"
)

func GetDriveItem(ctx context.Context, driveID string, itemID string) error {
	if driveID == "me" {
		fmt.Println("Error: drive_id must be the actual drive ID, cannot be 'me'")
		return nil
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	item, err := graph.GetDriveItem(ctx, c, driveID, itemID)
	if err != nil {
		return fmt.Errorf("failed to get drive item: %w", err)
	}

	fmt.Printf("\nDrive Item Details:\n")

	// Print item type and name
	if folder := item.GetFolder(); folder != nil {
		fmt.Printf("Type: 📁 Folder\n")
	} else {
		fmt.Printf("Type: 📄 File\n")
	}

	fmt.Printf("Name: %s\n", util.Deref(item.GetName()))
	fmt.Printf("ID: %s\n", util.Deref(item.GetId()))

	if parentRef := item.GetParentReference(); parentRef != nil {
		if parentID := parentRef.GetId(); parentID != nil {
			fmt.Printf("Parent ID: %s\n", util.Deref(parentID))
		}
		if path := parentRef.GetPath(); path != nil {
			fmt.Printf("Path: %s\n", util.Deref(path))
		}
	}

	if size := item.GetSize(); size != nil {
		fmt.Printf("Size: %d bytes\n", util.Deref(size))
	}

	if created := item.GetCreatedDateTime(); created != nil {
		fmt.Printf("Created: %s\n", created.Format(time.RFC3339))
	}

	if modified := item.GetLastModifiedDateTime(); modified != nil {
		fmt.Printf("Modified: %s\n", modified.Format(time.RFC3339))
	}

	if webUrl := item.GetWebUrl(); webUrl != nil {
		fmt.Printf("Web URL: %s\n", util.Deref(webUrl))
	}

	// Folder-specific information
	if folder := item.GetFolder(); folder != nil {
		fmt.Printf("Child Count: %d\n", util.Deref(folder.GetChildCount()))
	}

	// File-specific information
	if file := item.GetFile(); file != nil {
		if mimeType := file.GetMimeType(); mimeType != nil {
			fmt.Printf("MIME Type: %s\n", util.Deref(mimeType))
		}
		if hashes := file.GetHashes(); hashes != nil {
			if quickXor := hashes.GetQuickXorHash(); quickXor != nil {
				fmt.Printf("QuickXOR Hash: %s\n", util.Deref(quickXor))
			}
			if sha1 := hashes.GetSha1Hash(); sha1 != nil {
				fmt.Printf("SHA1 Hash: %s\n", util.Deref(sha1))
			}
		}
	}

	// Shared information
	if shared := item.GetShared(); shared != nil {
		fmt.Println("\nSharing Information:")
		if owner := shared.GetOwner(); owner != nil {
			if user := owner.GetUser(); user != nil {
				fmt.Printf("Shared by: %s\n", util.Deref(user.GetDisplayName()))
			}
		}
		if scope := shared.GetScope(); scope != nil {
			fmt.Printf("Sharing Scope: %s\n", util.Deref(scope))
		}
	}

	return nil
}
