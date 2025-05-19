package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	graph "github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
)

func DeleteGroupThread(ctx context.Context, groupID, threadID string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := graph.DeleteGroupThread(ctx, c, groupID, threadID); err != nil {
		return err
	}

	fmt.Printf("Group thread %s deleted successfully\n", threadID)
	return nil
}
