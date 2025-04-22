package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
)

func ListAttachments(ctx context.Context, messageID string) error {
	trueMessageID, err := id.GetOutlookID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	attachments, err := graph.ListAttachments(ctx, c, trueMessageID)
	if err != nil {
		return fmt.Errorf("failed to list attachments: %w", err)
	}

	for _, attachment := range attachments {
		fmt.Printf("ID: %s, Name: %s\n",
			*attachment.GetId(),
			*attachment.GetName())
	}

	if len(attachments) == 0 {
		fmt.Println("no attachments found")
	}

	return nil
}
