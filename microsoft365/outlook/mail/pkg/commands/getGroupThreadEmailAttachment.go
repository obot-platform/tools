package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
)

func GetGroupThreadEmailAttachment(ctx context.Context, groupID, threadID, emailID, attachmentID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	result, err := graph.GetGroupThreadMessageAttachment(ctx, c, groupID, threadID, emailID, attachmentID)
	if err != nil {
		return fmt.Errorf("failed to get group thread email attachment: %w", err)
	}

	fmt.Println(result)
	return nil
}
