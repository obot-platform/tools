package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
	md "github.com/JohannesKaufmann/html-to-markdown"
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

		// Fetch posts (individual emails/messages) inside the thread and then print them
		posts, err := graph.ListThreadMessages(ctx, c, groupID, threadID)
		if err != nil {
			return fmt.Errorf("failed to list thread messages: %w", err)
		}

		fmt.Println("\n✉️ Messages:")
		for i, post := range posts {
			messageID := util.Deref(post.GetId())
			fmt.Printf("📧 Message %d, ID: %s\n", i+1, messageID)

			// Check if sender information is available
			if post.GetFrom() != nil && post.GetFrom().GetEmailAddress() != nil {
				fmt.Printf("👤 From: %s <%s>\n",
					util.Deref(post.GetFrom().GetEmailAddress().GetName()),
					util.Deref(post.GetFrom().GetEmailAddress().GetAddress()),
				)
			} else {
				fmt.Println("👤 Sender: Unknown")
			}

			fmt.Printf("📅 Sent: %s\n", post.GetReceivedDateTime().String())

			// Print message body if available
			if post.GetBody() != nil && post.GetBody().GetContent() != nil {
				fmt.Println("📝 Message Body:")
				converter := md.NewConverter("", true, nil)
				bodyHTML := util.Deref(post.GetBody().GetContent())
				bodyMarkdown, err := converter.ConvertString(bodyHTML)
				if err != nil {
					return fmt.Errorf("failed to convert email body HTML to markdown: %w", err)
				}
				fmt.Println(bodyMarkdown)

			} else {
				fmt.Println("📭 (No content in this message)")
			}
			fmt.Println()
		}

		fmt.Println("\n")
	}
	return nil
} 