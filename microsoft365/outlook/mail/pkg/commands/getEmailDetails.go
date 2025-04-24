package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/printers"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/util"
)

func GetEmailDetails(ctx context.Context, emailID string) error {
	trueEmailID, err := id.GetOutlookID(ctx, emailID)
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}

	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	result, err := graph.GetMessageDetails(ctx, c, trueEmailID)
	if err != nil {
		return fmt.Errorf("failed to get email details: %w", err)
	}

	result.SetId(&trueEmailID)

	parentFolderID, err := id.GetOutlookID(ctx, util.Deref(result.GetParentFolderId()))
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}
	result.SetParentFolderId(&parentFolderID)

	if err := printers.PrintMessage(result, true); err != nil {
		return fmt.Errorf("failed to print message: %w", err)
	}
	return nil
}
