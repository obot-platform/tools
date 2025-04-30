package main

import (
	"fmt"
	"os"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

func main() {
	apiKey := os.Getenv("OBOT_DEEPSEEK_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("OBOT_DEEPSEEK_MODEL_PROVIDER_API_KEY is not set, credential must be provided on a per-request basis")
	}

	cfg := &proxy.Config{
		APIKey:               apiKey,
		PersonalAPIKeyHeader: "X-Obot-OBOT_DEEPSEEK_MODEL_PROVIDER_API_KEY",
		ListenPort:           os.Getenv("PORT"),
		BaseURL:              "https://api.deepseek.com/v1",
		RewriteModelsFn:      proxy.RewriteAllModelsWithUsage("llm"),
		Name:                 "DeepSeek",
	}

	if len(os.Args) > 1 && os.Args[1] == "validate" {
		if err := cfg.Validate("/tools/deepseek-model-provider/validate"); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := proxy.Run(cfg); err != nil {
		panic(err)
	}
}
