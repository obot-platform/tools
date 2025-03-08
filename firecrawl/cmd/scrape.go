package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mendableai/firecrawl-go"
)

func Scrape(ctx context.Context, apiKey, url string) (string, error) {
	// Parameters are already validated in main.go, so we can proceed directly
	
	// Initialize the FirecrawlApp
	apiUrl := "https://api.firecrawl.dev"
	app, err := firecrawl.NewFirecrawlApp(apiKey, apiUrl)
	if err != nil {
		return "", fmt.Errorf("failed to initialize FirecrawlApp: %w", err)
	}

	// Scrape with hardcoded format
	params := &firecrawl.ScrapeParams{
		Formats: []string{"markdown"},
	}

	scrapeResult, err := app.ScrapeURL(url, params)
	if err != nil {
		return "", fmt.Errorf("failed to scrape URL: %w", err)
	}

	resultJSON, err := json.MarshalIndent(scrapeResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal scrape result: %w", err)
	}

	return string(resultJSON), nil
}
