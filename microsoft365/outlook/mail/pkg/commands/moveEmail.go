package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/util"
)

func MoveEmail(ctx context.Context, emailID, destinationFolderID string) error {
	trueEmailID, err := id.GetOutlookID(ctx, emailID)
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}

	trueDestinationFolderID, err := id.GetOutlookID(ctx, destinationFolderID)
	if err != nil {
		return fmt.Errorf("failed to get destination folder ID: %w", err)
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	email, err := graph.MoveMessage(ctx, c, trueEmailID, trueDestinationFolderID)
	if err != nil {
		return fmt.Errorf("failed to move email: %w", err)
	}

	newEmailID, err := id.SetOutlookID(ctx, util.Deref(email.GetId()))
	if err != nil {
		return fmt.Errorf("failed to set outlook ID: %w", err)
	}

	fmt.Printf("Email moved successfully. New email ID: %s\n", newEmailID)
	return nil
}
