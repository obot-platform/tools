package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/obot-platform/tools/firecrawl/cmd"
)

func validateAPIKey() string {
	apiKey := strings.TrimSpace(os.Getenv("FIRECRAWL_API_KEY"))
	if apiKey == "" {
		exitWithError("API key is required")
	}
	return apiKey
}

func validateRequiredParam(value, name string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		exitWithError(fmt.Sprintf("%s is required", name))
	}
	return value
}

func exitWithError(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gptscript-go-tool <command>")
		os.Exit(1)
	}
	command := os.Args[1]

	ctx := context.Background()

	apiKey := validateAPIKey()

	switch command {
	case "scrapeUrl":
		urls := validateRequiredParam(os.Getenv("URLS"), "URLS")

		// clean and deduplicate urls
		urlMap := make(map[string]struct{})
		for _, url := range strings.Split(urls, ",") {
			url = strings.TrimSpace(url)
			if _, ok := urlMap[url]; ok || url == "" {
				continue
			}
			urlMap[url] = struct{}{}
		}

		// Scrape each URL concurrently
		results := make(chan string, len(urlMap))
		wg := sync.WaitGroup{}
		wg.Add(len(urlMap))

		for url := range urlMap {
			go func(url string) {
				defer wg.Done()
				content, err := cmd.Scrape(ctx, apiKey, url)
				if err != nil {
					content = fmt.Sprintf("ERROR: failed to scrape URL %q: %v", url, err)
				}
				results <- fmt.Sprintf("!==== Start Contents of URL %q ====!\n%s\n!==== End Contents of URL %q ====!\n", url, content, url)
			}(url)
		}

		wg.Wait()
		close(results)

		for content := range results {
			fmt.Println(content)
		}

	default:
		exitWithError(fmt.Sprintf("unknown command: %s", command))
	}
}
