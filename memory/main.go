package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/obot-platform/obot/apiclient"
	"github.com/obot-platform/obot/pkg/system"
)

var (
	url            = os.Getenv("OBOT_SERVER_URL")
	token          = os.Getenv("OBOT_TOKEN")
	threadID       = os.Getenv("OBOT_THREAD_ID")
	parentThreadID = os.Getenv("OBOT_PROJECT_ID")
	assistantID    = os.Getenv("OBOT_AGENT_ID")
	content        = os.Getenv("CONTENT")
	memoryID       = os.Getenv("MEMORY_ID")
)

func main() {
	ctx := context.Background()
	if err := mainErr(ctx); err != nil {
		slog.Error("error", "err", err)
		os.Exit(1)
	}
}

func create(ctx context.Context, c *apiclient.Client, projectID, content string) error {
	if content == "" {
		return fmt.Errorf("missing content to remember")
	}

	memory, err := c.CreateMemory(ctx, assistantID, projectID, content)
	if err != nil {
		return fmt.Errorf("failed to create memory: %v", err)
	}

	fmt.Printf("memory %q created", memory.ID)

	return nil
}

func update(ctx context.Context, c *apiclient.Client, projectID, content string) error {
	if memoryID == "" {
		return fmt.Errorf("missing memory_id")
	}

	if content == "" {
		return fmt.Errorf("missing content to remember")
	}

	memory, err := c.UpdateMemory(ctx, assistantID, projectID, memoryID, content)
	if err != nil {
		return fmt.Errorf("failed to update memory: %v", err)
	}

	fmt.Printf("memory %q updated", memory.ID)

	return nil
}

func delete(ctx context.Context, c *apiclient.Client, projectID string) error {
	if memoryID == "" {
		return fmt.Errorf("missing memory_id")
	}

	if err := c.DeleteMemory(ctx, assistantID, projectID, memoryID); err != nil {
		return fmt.Errorf("failed to delete memory: %v", err)
	}

	fmt.Printf("memory %q deleted", memoryID)

	return nil
}

func list(ctx context.Context, c *apiclient.Client, projectID string) error {
	result, err := c.ListMemories(ctx, assistantID, projectID)
	if err != nil && !strings.Contains(err.Error(), "404") {
		return fmt.Errorf("failed to list memories: %v", err)
	}

	var sb strings.Builder
	sb.WriteString("Below are memories for you to reference when crafting responses to the user:\n")
	sb.WriteString("<MEMORIES>\n")

	// Add header row
	sb.WriteString("memory_id, content\n")

	// Add each memory as a CSV row
	if result != nil && len(result.Items) > 0 {
		for _, memory := range result.Items {
			sb.WriteString(fmt.Sprintf("%s, %s\n", memory.ID, memory.Content))
		}
	}

	// Add closing tag and output
	sb.WriteString("</MEMORIES>")
	fmt.Print(sb.String())

	return nil
}

func mainErr(ctx context.Context) error {
	var projectID string
	if parentThreadID != "" {
		projectID = strings.Replace(parentThreadID, system.ThreadPrefix, system.ProjectPrefix, 1)
	} else if threadID != "" {
		projectID = strings.Replace(threadID, system.ThreadPrefix, system.ProjectPrefix, 1)
	}

	if projectID == "" {
		return fmt.Errorf("missing project id")
	}

	if len(os.Args) == 1 {
		fmt.Printf("incorrect usage: %s [create|update|delete|list]\n", os.Args[0])
		return nil
	}

	if url == "" {
		url = "http://localhost:8080/api"
	} else {
		url += "/api"
	}

	client := &apiclient.Client{
		BaseURL: url,
		Token:   token,
	}

	switch os.Args[1] {
	case "create":
		return create(ctx, client, projectID, content)
	case "update":
		return update(ctx, client, projectID, content)
	case "delete":
		return delete(ctx, client, projectID)
	case "list":
		return list(ctx, client, projectID)
	}

	return nil
}
