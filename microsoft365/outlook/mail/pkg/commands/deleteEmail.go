package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
)

func DeleteEmail(ctx context.Context, emailID string) error {
	trueEmailID, err := id.GetOutlookID(ctx, emailID)
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := graph.DeleteMessage(ctx, c, trueEmailID); err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}

	fmt.Println("Email deleted successfully.")
	return nil
}
