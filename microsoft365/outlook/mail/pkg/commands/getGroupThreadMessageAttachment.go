package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
)

func GetGroupThreadMessageAttachment(ctx context.Context, groupID, threadID, postID, attachmentID string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	result, err := graph.GetGroupThreadMessageAttachment(ctx, c, groupID, threadID, postID, attachmentID)
	if err != nil {
		return fmt.Errorf("failed to get attachment: %w", err)
	}

	fmt.Println(result)
	return nil
}
