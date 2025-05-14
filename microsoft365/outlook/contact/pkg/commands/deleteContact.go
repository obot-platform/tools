package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/graph"
)

func DeleteContact(ctx context.Context, contactID string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	err = graph.DeleteContact(ctx, c, contactID)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	fmt.Printf("Contact %s deleted successfully\n", contactID)

	return nil
}
