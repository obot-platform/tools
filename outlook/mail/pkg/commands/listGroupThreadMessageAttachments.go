package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
)

func ListGroupThreadMessageAttachments(ctx context.Context, groupID, threadID, postID string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	attachments, err := graph.ListGroupThreadMessageAttachments(ctx, c, groupID, threadID, postID)
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
