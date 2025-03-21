package googlecustomsearchengine

import (
	"context"
	"os"

	googlecustomsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

type GoogleCSE struct {
	apiKey string
	cseID  string
}

func NewGoogleCSEFromEnv() *GoogleCSE {
	return &GoogleCSE{
		apiKey: os.Getenv("GOOGLE_CSE_API_KEY"),
		cseID:  os.Getenv("GOOGLE_CSE_ID"),
	}
}

type Result struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

func (g *GoogleCSE) Search(ctx context.Context, query string) ([]Result, error) {
	client, err := googlecustomsearch.NewService(ctx, option.WithAPIKey(g.apiKey))
	if err != nil {
		return nil, err
	}

	resp, err := client.Cse.List().Cx(g.cseID).Q(query).Do()
	if err != nil {
		return nil, err
	}

	results := make([]Result, len(resp.Items))

	for i, item := range resp.Items {
		results[i] = Result{
			Title:   item.Title,
			Link:    item.Link,
			Snippet: item.Snippet,
		}
	}

	return results, nil
}
