package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
)

func GetMyEmailAddress(ctx context.Context) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	user, err := graph.GetMe(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to get me: %w", err)
	}

	email := user.GetMail()
	if email == nil {
		return fmt.Errorf("failed to get email address")
	}

	fmt.Printf("Current user email address is %s\n", *email)
	return nil
}
