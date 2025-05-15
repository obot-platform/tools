package commands

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/client"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/global"
	"github.com/obot-platform/tools/microsoft365/onedrive/pkg/graph"
)

func AddPermission(ctx context.Context, driveID string, itemID string, emails string, role string, message string, password string, expirationDateTime string) error {
	if driveID == "me" {
		return fmt.Errorf("drive_id must be the actual drive ID, cannot be 'me'")
	}

	var emailList []string
	if emails != "" {
		for _, email := range strings.Split(strings.ReplaceAll(strings.TrimSpace(emails), " ", ""), ",") {
			if email != "" {
				emailList = append(emailList, email)
			}
		}
	}

	validRoles := []string{"read", "write"}
	if !slices.Contains(validRoles, role) {
		return fmt.Errorf("invalid role: %s, valid roles are: %s", role, strings.Join(validRoles, ", "))
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	_, err = graph.AddDriveItemPermission(ctx, c, driveID, itemID, emailList, role, message, password, expirationDateTime)
	if err != nil {
		return fmt.Errorf("failed to add permissions: %w", err)
	}

	fmt.Printf("Successfully added permissions for item %s in drive %s\n", itemID, driveID)
	fmt.Printf("Permission has been added to the following emails: %s\n", strings.Join(emailList, ", "))
	return nil
}
