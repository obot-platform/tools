package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/obot-platform/obot/apiclient"
)

var (
	url        = os.Getenv("OBOT_SERVER_URL")
	token      = os.Getenv("OBOT_TOKEN")
	id         = os.Getenv("ID")
	threadID   = os.Getenv("OBOT_THREAD_ID")
	args       = os.Getenv("ARGS")
	definition = os.Getenv("DEFINITION")
)

func main() {
	ctx := context.Background()
	if err := mainErr(ctx); err != nil {
		slog.Error("error", "err", err)
		os.Exit(1)
	}
}

func list(ctx context.Context, c *apiclient.Client) error {
	result, err := c.ListTasks(ctx, apiclient.ListTasksOptions{
		ThreadID: threadID,
	})
	if err != nil {
		return fmt.Errorf("list tasks: %v", err)
	}

	if len(result.Items) == 0 {
		fmt.Printf("no tasks found\n")
		return nil
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func runs(ctx context.Context, c *apiclient.Client, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("missing task ID")
	}
	result, err := c.ListTaskRuns(ctx, taskID, apiclient.ListTaskRunsOptions{
		ThreadID: threadID,
	})
	if err != nil {
		return fmt.Errorf("list runs: %v", err)
	}

	if len(result.Items) == 0 {
		fmt.Printf("no runs found\n")
		return nil
	}

	return json.NewEncoder(os.Stdout).Encode(result)
}

func mainErr(ctx context.Context) error {
	if len(os.Args) == 1 {
		fmt.Printf("incorrect usage: %s [list|run|create|list-runs]\n", os.Args[0])
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
	case "list":
		return list(ctx, client)
	case "list-runs":
		return runs(ctx, client, id)
	}

	return nil
}
