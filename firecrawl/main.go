package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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

	var (
		result string
		err    error
		ctx    = context.Background()
	)

	apiKey := validateAPIKey()

	switch command {
	case "scrapeUrl":
		url := validateRequiredParam(os.Getenv("URL"), "URL")
		result, err = cmd.Scrape(ctx, apiKey, url)
		
	default:
		exitWithError(fmt.Sprintf("unknown command: %s", command))
	}

	if err != nil {
		exitWithError(err.Error())
	}

	fmt.Print(result)
}
