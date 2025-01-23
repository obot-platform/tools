package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

func cleanURL(endpoint string) string {
	return strings.TrimRight(endpoint, "/")
}

func main() {
	apiKey := os.Getenv("OBOT_VLLM_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OBOT_VLLM_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	endpoint := os.Getenv("OBOT_VLLM_MODEL_PROVIDER_ENDPOINT")
	if endpoint == "" {
		fmt.Fprintln(os.Stderr, "OBOT_VLLM_MODEL_PROVIDER_ENDPOINT environment variable not set")
		os.Exit(1)
	}

	endpoint = cleanURL(endpoint)
	u, err := url.Parse(endpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid endpoint URL %q: %v\n", endpoint, err)
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
		APIKey:          apiKey,
		Port:            os.Getenv("PORT"),
		UpstreamHost:    u.Host,
		UseTLS:          u.Scheme == "https",
		RewriteModelsFn: proxy.RewriteAllModelsWithUsage("llm"),
		Name:            "vLLM",
	}

	if len(os.Args) > 1 && os.Args[1] == "validate" {
		if err := cfg.Validate("/tools/vllm-model-provider/validate"); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := proxy.Run(cfg); err != nil {
		panic(err)
	}
}
