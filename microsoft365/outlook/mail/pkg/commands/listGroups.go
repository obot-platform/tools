package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/util"
)

func ListGroups(ctx context.Context) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	groups, err := graph.ListGroups(ctx, c)
	if err != nil {
		userType, getUserTypeErr := graph.GetUserType(ctx, c)
		if getUserTypeErr == nil {
			if strings.EqualFold(userType, "Guest") || strings.EqualFold(userType, "Personal") { // Guest or Personal accounts don't have enough permissions to list groups
				fmt.Printf("User has type '%s', which does not have enough permissions to list groups.\n", userType)
				return nil
			}
		}
		return fmt.Errorf("failed to list groups: %w", err)
	}

	if len(groups) == 0 {
		fmt.Println("No groups found")
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	var elements []gptscript.DatasetElement
	for _, group := range groups {
		elements = append(elements, gptscript.DatasetElement{
			DatasetElementMeta: gptscript.DatasetElementMeta{
				Name:        util.Deref(group.GetDisplayName()),
				Description: util.Deref(group.GetDescription()),
			},
			Contents: fmt.Sprintf("ID: %s\nName: %s\nDescription: %s\nMail: %s\n",
				util.Deref(group.GetId()),
				util.Deref(group.GetDisplayName()),
				util.Deref(group.GetDescription()),
				util.Deref(group.GetMail()),
			),
		})
	}

	if len(elements) == 0 {
		return nil
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	datasetID, err := gptscriptClient.CreateDatasetWithElements(ctx, elements, gptscript.DatasetOptions{
		Name:        "outlook_groups",
		Description: "Outlook groups",
	})
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d groups\n", datasetID, len(groups))
	return nil
}
