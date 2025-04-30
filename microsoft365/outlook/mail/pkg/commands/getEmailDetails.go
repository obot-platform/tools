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

func GetEmailDetails(ctx context.Context, emailID, groupID, threadID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	if groupID == "" { // Personal mailbox
		trueEmailID, err := id.GetOutlookID(ctx, emailID)
		if err != nil {
			return fmt.Errorf("failed to get outlook ID: %w", err)
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
	// Group mailbox
	result, err := graph.GetThreadMessage(ctx, c, groupID, threadID, emailID)
	if err != nil {
		return fmt.Errorf("failed to get group mailbox email details: %w", err)
	}

	fmt.Printf("Message Details:\n")
	fmt.Printf("Body Content Type: %s\n", util.Deref(result.GetBody().GetContentType()))
	fmt.Printf("Body Content: %s\n", util.Deref(result.GetBody().GetContent()))
	fmt.Printf("Received Date Time: %s\n", result.GetReceivedDateTime().Format("2006-01-02 15:04:05"))
	fmt.Printf("Has Attachments: %v\n", result.GetHasAttachments())

	if from := result.GetFrom(); from != nil {
		if email := from.GetEmailAddress(); email != nil {
			fmt.Printf("From Name: %s\n", util.Deref(email.GetName()))
			fmt.Printf("From Address: %s\n", util.Deref(email.GetAddress()))
		}
	}

	if sender := result.GetSender(); sender != nil {
		if email := sender.GetEmailAddress(); email != nil {
			fmt.Printf("Sender Name: %s\n", util.Deref(email.GetName()))
			fmt.Printf("Sender Address: %s\n", util.Deref(email.GetAddress()))
		}
	}

	return nil
}
