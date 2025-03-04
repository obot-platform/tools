package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	googlecustomsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

func exitError(err error) {
	if err != nil {
		fmt.Printf("google custom search failed: %v\n", err)
		os.Exit(1)
	}
}

type result struct {
	Title   string
	Link    string
	Snippet string
}

func main() {
	ctx := context.Background()
	res, err := search(ctx)
	exitError(err)
	fmt.Println(res)
}

func search(ctx context.Context) (string, error) {
	apiKey := os.Getenv("GOOGLE_CSE_API_KEY")
	client, err := googlecustomsearch.NewService(ctx, option.WithAPIKey(apiKey))
	exitError(err)

	cseID := os.Getenv("GOOGLE_CSE_ID")
	query := os.Getenv("QUERY")

	resp, err := client.Cse.List().Cx(cseID).Q(query).Do()
	exitError(err)

	results := make([]result, len(resp.Items))

	for i, item := range resp.Items {
		results[i] = result{
			Title:   item.Title,
			Link:    item.Link,
			Snippet: item.Snippet,
		}
	}

	resJSON, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	return string(resJSON), nil
}
