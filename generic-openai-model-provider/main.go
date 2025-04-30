package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

func main() {
	isValidate := len(os.Args) > 1 && os.Args[1] == "validate"

	baseURL := os.Getenv("OBOT_GENERIC_OPENAI_MODEL_PROVIDER_BASE_URL")
	if baseURL == "" {
		fmt.Println("OBOT_GENERIC_OPENAI_MODEL_PROVIDER_BASE_URL environment variable not set, credential must be provided on a per-request basis")
	}

	u, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid baseURL %q: %v\n", baseURL, err)
		fmt.Printf("{ \"error\": \"Invalid BaseURL: %v\" }\n", err)
		os.Exit(1)
	}

	if u.Scheme == "" {
		if u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" {
			u.Scheme = "http"
		} else {
			u.Scheme = "https"
		}
	}

	cfg := &proxy.Config{
		APIKey:               os.Getenv("OBOT_GENERIC_OPENAI_MODEL_PROVIDER_API_KEY"), // optional, as e.g. Ollama doesn't require an API key
		PersonalAPIKeyHeader: "X-Obot-OBOT_GENERIC_OPENAI_MODEL_PROVIDER_API_KEY",
		ListenPort:           os.Getenv("PORT"),
		BaseURL:              u.String(),
		RewriteModelsFn:      proxy.RewriteAllModelsWithUsage("llm"),
		Name:                 "Generic OpenAI",
	}

	if err := cfg.Validate("/tools/generic-openai-model-provider/validate"); err != nil {
		os.Exit(1)
	}

	if isValidate {
		return
	}

	if err := proxy.Run(cfg); err != nil {
		fmt.Printf("failed to run generic-openai-model-provider proxy: %v\n", err)
		os.Exit(1)
	}
}
