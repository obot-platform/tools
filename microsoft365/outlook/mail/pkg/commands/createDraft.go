package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/util"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func CreateDraft(ctx context.Context, info graph.DraftInfo) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	var draft models.Messageable
	if info.ReplyToMessageID != "" {
		draft, err = graph.CreateDraftReply(ctx, c, info)
		if err != nil {
			return fmt.Errorf("failed to create draft reply: %w", err)
		}
	} else {
		draft, err = graph.CreateDraft(ctx, c, info)
		if err != nil {
			return fmt.Errorf("failed to create draft: %w", err)
		}
	}

	// Get numerical ID for the draft
	draftID, err := id.SetOutlookID(ctx, util.Deref(draft.GetId()))
	if err != nil {
		return fmt.Errorf("failed to set draft ID: %w", err)
	}

	fmt.Printf("Draft created successfully. Draft ID: %s\n", draftID)
	return nil
}
