package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
)

func CreateGroupThreadMessage(ctx context.Context, groupID string, info graph.DraftInfo) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	threads, err := graph.CreateGroupThreadMessage(ctx, c, groupID, info)
	if err != nil {
		return fmt.Errorf("failed to create group thread message: %w", err)
	}

	fmt.Println("Group thread message created successfully, thread ID:", util.Deref(threads.GetId()))
	return nil
}
