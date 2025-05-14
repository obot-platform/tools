package commands

import (
	"context"
	"fmt"

	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/util"
)

func GetContact(ctx context.Context, contactID string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	contact, err := graph.GetContact(ctx, c, contactID)
	if err != nil {
		return fmt.Errorf("failed to get contact: %w", err)
	}

	// Print contact information
	fmt.Println("Contact Details:")
	fmt.Printf("ID: %s\n", util.Deref(contact.GetId()))

	if contact.GetGivenName() != nil {
		fmt.Printf("First Name: %s\n", util.Deref(contact.GetGivenName()))
	}

	if contact.GetSurname() != nil {
		fmt.Printf("Last Name: %s\n", util.Deref(contact.GetSurname()))
	}

	if contact.GetDisplayName() != nil {
		fmt.Printf("Display Name: %s\n", util.Deref(contact.GetDisplayName()))
	}

	if emailAddresses := contact.GetEmailAddresses(); emailAddresses != nil && len(emailAddresses) > 0 {
		fmt.Println("Email Addresses:")
		for i, email := range emailAddresses {
			if email.GetAddress() != nil {
				emailAddress := util.Deref(email.GetAddress())
				fmt.Printf("  %d. %s\n", i+1, emailAddress)
			}
		}
	}

	if phones := contact.GetBusinessPhones(); phones != nil && len(phones) > 0 {
		fmt.Println("Business Phones:")
		for i, phone := range phones {
			fmt.Printf("  %d. %s\n", i+1, phone)
		}
	}

	return nil
}
