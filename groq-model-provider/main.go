package main

import (
	"fmt"
	"os"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

func main() {
	apiKey := os.Getenv("OBOT_GROQ_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("OBOT_GROQ_MODEL_PROVIDER_API_KEY is not set, credential must be provided on a per-request basis")
	}

	cfg := &proxy.Config{
		APIKey:               apiKey,
		PersonalAPIKeyHeader: "X-Obot-OBOT_GROQ_MODEL_PROVIDER_API_KEY",
		ListenPort:           os.Getenv("PORT"),
		BaseURL:              "https://api.groq.com/openai/v1",
		RewriteModelsFn:      proxy.RewriteAllModelsWithUsage("llm"),
		Name:                 "Groq",
	}

	if len(os.Args) > 1 && os.Args[1] == "validate" {
		if err := cfg.Validate("/tools/groq-model-provider/validate"); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := proxy.Run(cfg); err != nil {
		panic(err)
	}
}
