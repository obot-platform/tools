package commands

import (
	"context"
	"fmt"
	"strconv"

	// "github.com/gptscript-ai/go-gptscript"
	// "github.com/gptscript-ai/tools/outlook/common/id"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	// "github.com/gptscript-ai/tools/outlook/mail/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
	// "github.com/microsoftgraph/msgraph-sdk-go/models"
)



func ListGroupThreads(ctx context.Context, groupID, start, end, limit string) error {
	var (
		limitInt int = 100
		err      error
	)
	if limit != "" {
		limitInt, err = strconv.Atoi(limit)
		if err != nil {
			return fmt.Errorf("failed to parse limit: %w", err)
		}
		if limitInt < 1 {
			return fmt.Errorf("limit must be a positive integer")
		}
	}

	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	threads, err := graph.ListGroupThreads(ctx, c, groupID, start, end, limitInt)
	if err != nil {
		return fmt.Errorf("failed to list group threads: %w", err)
	}

	for _, thread := range threads {
		fmt.Println("==========================================")
		threadID := util.Deref(thread.GetId())
		fmt.Printf("📩 Thread ID: %s\n", threadID)
		if thread.GetTopic() != nil {
			fmt.Printf("📌 Subject: %s\n", util.Deref(thread.GetTopic()))
		} else {
			fmt.Println("📌 Subject: (No Subject)")
		}
		fmt.Printf("📅 Last Delivered: %s\n", thread.GetLastDeliveredDateTime().String())

		// Print unique senders
		senders := thread.GetUniqueSenders()
		fmt.Print("👥 Unique Senders: ")
		for _, sender := range senders {
			fmt.Printf("%s, ", sender) 
		}
		fmt.Println()

		// Fetch posts (individual emails/messages) inside the thread
		
		graph.PrintThreadMessages(ctx, c, groupID, threadID)

		fmt.Println("==========================================")
	}
	return nil
} 