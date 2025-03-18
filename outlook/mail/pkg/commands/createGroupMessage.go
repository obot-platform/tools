package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/common/id"
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

	// Get numerical ID for the draft
	threadID, err := id.SetOutlookID(ctx, util.Deref(threads.GetId()))
	if err != nil {
		return fmt.Errorf("failed to set draft ID: %w", err)
	}

	fmt.Printf("Group thread message created successfully. Thread ID: %s\n", threadID)
	return nil
}
