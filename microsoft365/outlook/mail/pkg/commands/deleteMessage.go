package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
)

func DeleteMessage(ctx context.Context, messageID string) error {
	trueMessageID, err := id.GetOutlookID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message ID: %w", err)
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := graph.DeleteMessage(ctx, c, trueMessageID); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	fmt.Println("Message deleted successfully")
	return nil
}
