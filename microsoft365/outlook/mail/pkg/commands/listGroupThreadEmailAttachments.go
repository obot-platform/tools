package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
)

func ListGroupThreadEmailAttachments(ctx context.Context, groupID, threadID, emailID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	attachments, err := graph.ListGroupThreadMessageAttachments(ctx, c, groupID, threadID, emailID)
	if err != nil {
		return fmt.Errorf("failed to list group thread email attachments: %w", err)
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
