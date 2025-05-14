package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/contact/pkg/util"
)

func ListContacts(ctx context.Context) error {
	// TODO: support filtering. see https://learn.microsoft.com/en-us/graph/query-parameters?tabs=http

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	contacts, err := graph.ListAllContacts(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list contacts: %w", err)
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	var elements []gptscript.DatasetElement
	for _, contact := range contacts {
		// Extract contact details
		id := util.Deref(contact.GetId())
		name := util.Deref(contact.GetDisplayName())
		itemStr := fmt.Sprintf("Contact ID: %s, Name: %s", id, name)
		if emails := contact.GetEmailAddresses(); len(emails) > 0 && emails[0] != nil {
			for i, email := range emails {
				itemStr += fmt.Sprintf(", Email %d: %s", i+1, util.Deref(email.GetAddress()))
			}
		}
		if phones := contact.GetBusinessPhones(); len(phones) > 0 {
			for i, phone := range phones {
				itemStr += fmt.Sprintf(", Phone %d: %s", i+1, phone)
			}
		}

		elements = append(elements, gptscript.DatasetElement{
			DatasetElementMeta: gptscript.DatasetElementMeta{
				Name:        id,
				Description: "Contact Person: " + name,
			},
			Contents: itemStr,
		})
	}

	if len(elements) == 0 {
		fmt.Println("No contacts found")
		return nil
	}

	datasetID, err := gptscriptClient.CreateDatasetWithElements(ctx, elements, gptscript.DatasetOptions{
		Name:        "outlook_contacts",
		Description: "Outlook contacts",
	})
	if err != nil {
		return fmt.Errorf("failed to create dataset with elements: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d contacts items\n", datasetID, len(contacts))

	return nil
}
