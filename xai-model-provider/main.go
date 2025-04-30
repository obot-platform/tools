package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

// RewriteGrokModels marks only Grok models as LLMs
func RewriteGrokModels(resp *http.Response) error {
	rewriteFn := proxy.RewriteAllModelsWithUsage("llm", func(modelID string) bool {
		return strings.HasPrefix(modelID, "grok-")
	})
	return rewriteFn(resp)
}

func main() {
	apiKey := os.Getenv("OBOT_XAI_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("OBOT_XAI_MODEL_PROVIDER_API_KEY is not set, credential must be provided on a per-request basis")
	}

	cfg := &proxy.Config{
		APIKey:               apiKey,
		PersonalAPIKeyHeader: "X-Obot-OBOT_XAI_MODEL_PROVIDER_API_KEY",
		ListenPort:           os.Getenv("PORT"),
		BaseURL:              "https://api.x.ai/v1",
		RewriteModelsFn:      RewriteGrokModels,
		Name:                 "xAI",
	}

	if len(os.Args) > 1 && os.Args[1] == "validate" {
		if err := cfg.Validate("/tools/xai-model-provider/validate"); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := proxy.Run(cfg); err != nil {
		panic(err)
	}
}
