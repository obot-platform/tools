package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
)

func GetAttachment(ctx context.Context, messageID, attachmentID string) error {
	trueMessageID, err := id.GetOutlookID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Get attachment as a Parsable object
	requestInfo, err := c.Me().Messages().ByMessageId(trueMessageID).Attachments().ByAttachmentId(attachmentID).ToGetRequestInformation(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create request info: %w", err)
	}

	result, err := graph.GetAttachmentContent(ctx, c, requestInfo)
	if err != nil {
		return fmt.Errorf("failed to get attachment content: %w", err)
	}

	fmt.Println(result)
	return nil
}
