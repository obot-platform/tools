package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/graph"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/util"
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

	threads, err := graph.ListGroupThreads(ctx, c, groupID, limitInt)
	if err != nil {
		return fmt.Errorf("failed to list group threads: %w", err)
	}

	if len(threads) == 0 {
		fmt.Println("No threads found")
		return nil
	}

	if start != "" || end != "" {
		threads, err = filterThreadsByTimeFrame(threads, start, end)
		if err != nil {
			return fmt.Errorf("failed to filter threads by time frame: %w", err)
		}

		if len(threads) == 0 {
			fmt.Println("No threads found within specified time frame")
			return nil
		}
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	var (
		elements  []gptscript.DatasetElement
		converter = md.NewConverter("", true, nil)
	)

	for _, thread := range threads {
		threadID := util.Deref(thread.GetId())
		threadSubject := util.Deref(thread.GetTopic())
		if threadSubject == "" {
			threadSubject = "(No Subject)"
		}

		// Build thread content string
		var threadContent string
		threadContent += fmt.Sprintf("Thread ID: %s\n", threadID)
		threadContent += fmt.Sprintf("Subject: %s\n", threadSubject)
		threadContent += fmt.Sprintf("Last Delivered: %s\n", thread.GetLastDeliveredDateTime().String())

		// Add unique senders
		senders := thread.GetUniqueSenders()
		threadContent += "Unique Senders: " + strings.Join(senders, ", ") + "\n\n"

		// Fetch and add messages
		posts, err := graph.ListThreadMessages(ctx, c, groupID, threadID)
		if err != nil {
			return fmt.Errorf("failed to list thread messages: %w", err)
		}

		threadContent += "Messages:\n"
		for i, post := range posts {
			messageID := util.Deref(post.GetId())
			threadContent += fmt.Sprintf("\nMessage %d (ID: %s)\n", i+1, messageID)

			if post.GetFrom() != nil && post.GetFrom().GetEmailAddress() != nil {
				threadContent += fmt.Sprintf("From: %s <%s>\n",
					util.Deref(post.GetFrom().GetEmailAddress().GetName()),
					util.Deref(post.GetFrom().GetEmailAddress().GetAddress()),
				)
			} else {
				threadContent += "From: Unknown\n"
			}

			threadContent += fmt.Sprintf("Sent: %s\n", post.GetReceivedDateTime().String())

			if post.GetBody() != nil && post.GetBody().GetContent() != nil {
				bodyHTML := util.Deref(post.GetBody().GetContent())
				bodyMarkdown, err := converter.ConvertString(bodyHTML)
				if err != nil {
					return fmt.Errorf("failed to convert email body HTML to markdown: %w", err)
				}
				threadContent += fmt.Sprintf("Body:\n%s\n", bodyMarkdown)
			} else {
				threadContent += "(No content in this message)\n"
			}
		}

		elements = append(elements, gptscript.DatasetElement{
			DatasetElementMeta: gptscript.DatasetElementMeta{
				Name:        threadID,
				Description: threadSubject,
			},
			Contents: threadContent,
		})
	}

	datasetID, err := gptscriptClient.CreateDatasetWithElements(ctx, elements, gptscript.DatasetOptions{
		Name:        fmt.Sprintf("outlook_group_%s_threads", groupID),
		Description: fmt.Sprintf("Threads from Outlook group %s", groupID),
	})
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d threads\n", datasetID, len(elements))
	return nil
}

func filterThreadsByTimeFrame(threads []models.ConversationThreadable, start, end string) ([]models.ConversationThreadable, error) {
	var (
		startTime time.Time
		endTime   time.Time
		err       error
	)
	if start != "" {
		startTime, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start date: %w", err)
		}
	}
	if end != "" {
		endTime, err = time.Parse(time.RFC3339, end)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end date: %w", err)
		}
	}

	var filteredThreads []models.ConversationThreadable
	for _, thread := range threads {
		if start != "" && thread.GetLastDeliveredDateTime().Before(startTime) {
			continue
		}
		if end != "" && thread.GetLastDeliveredDateTime().After(endTime) {
			continue
		}
		filteredThreads = append(filteredThreads, thread)
	}

	return filteredThreads, nil
}
