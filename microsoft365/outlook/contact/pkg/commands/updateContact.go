package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/util"
)

func UpdateContact(ctx context.Context, contactID, givenName, surname, emailAddress, businessPhone string) error {

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	contact, err := graph.UpdateContact(ctx, c, contactID, givenName, surname, emailAddress, businessPhone)
	if err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
	}

	fmt.Printf("Contact updated successfully. ID: %s\n", util.Deref(contact.GetId()))

	return nil
}
