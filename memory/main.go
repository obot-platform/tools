package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/obot-platform/obot/apiclient"
	"github.com/obot-platform/obot/apiclient/types"
	"github.com/obot-platform/obot/pkg/system"
)

var (
	url            = os.Getenv("OBOT_SERVER_URL")
	token          = os.Getenv("OBOT_TOKEN")
	threadID       = os.Getenv("OBOT_THREAD_ID")
	parentThreadID = os.Getenv("OBOT_PROJECT_ID")
	assistantID    = os.Getenv("OBOT_AGENT_ID")
	content        = os.Getenv("CONTENT")
)

func main() {
	ctx := context.Background()
	if err := mainErr(ctx); err != nil {
		slog.Error("error", "err", err)
		os.Exit(1)
	}
}

func add(ctx context.Context, c *apiclient.Client, projectID, content string) error {
	if content == "" {
		return fmt.Errorf("missing content to remember")
	}

	if _, err := c.AddMemories(ctx, assistantID, projectID, types.Memory{
		Content: content,
	}); err != nil {
		return fmt.Errorf("failed to add memory: %v", err)
	}

	fmt.Printf("memory added")

	return nil
}

func list(ctx context.Context, c *apiclient.Client, projectID string) error {
	result, err := c.GetMemories(ctx, assistantID, projectID)
	if err != nil && !strings.Contains(err.Error(), "404") {
		return fmt.Errorf("failed to list memories: %v", err)
	}

	var sb strings.Builder
	sb.WriteString("Below are memories for you to reference when crafting responses to the user:\n")
	sb.WriteString("<MEMORIES>")

	if result != nil && len(result.Memories) > 0 {
		// Extract memory contents
		contents := make([]string, 0, len(result.Memories))
		for _, memory := range result.Memories {
			contents = append(contents, memory.Content)
		}

		// Marshal memory contents to JSON
		jsonData, err := json.Marshal(contents)
		if err != nil {
			return fmt.Errorf("failed to marshal memory contents: %v", err)
		}

		// Append JSON data to string builder
		sb.Write(jsonData)
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
		fmt.Printf("incorrect usage: %s [add|list]\n", os.Args[0])
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
	case "add":
		return add(ctx, client, projectID, content)
	case "list":
		return list(ctx, client, projectID)
	}

	return nil
}
