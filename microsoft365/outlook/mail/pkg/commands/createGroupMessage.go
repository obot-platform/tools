package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/util"
)

func CreateGroupThreadMessage(ctx context.Context, groupID, replyToThreadID string, info graph.DraftInfo) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if replyToThreadID != "" { // reply to a thread
		err = graph.ReplyToGroupThreadMessage(ctx, c, groupID, replyToThreadID, info)
		if err != nil {
			return fmt.Errorf("failed to reply to group thread message: %w", err)
		}
		fmt.Println("Group thread message replied to successfully")
		return nil
	} else { // create a new thread
		threads, err := graph.CreateGroupThreadMessage(ctx, c, groupID, info)
		if err != nil {
			return fmt.Errorf("failed to create group thread message: %w", err)
		}

		fmt.Println("Group thread message created successfully, thread ID:", util.Deref(threads.GetId()))
		return nil
	}
}
