package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
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

	for _, group := range groups {
		fmt.Printf("ID: %s\nName: %s\nDescription: %s\nMail: %s\n\n",
			util.Deref(group.GetId()),
			util.Deref(group.GetDisplayName()),
			util.Deref(group.GetDescription()),
			util.Deref(group.GetMail()),
		)
	}

	return nil
}