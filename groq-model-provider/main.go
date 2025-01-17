package main

import (
	"fmt"
	"os"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

func main() {
	apiKey := os.Getenv("OBOT_GROQ_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OBOT_GROQ_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	cfg := &proxy.Config{
		APIKey:          apiKey,
		Port:            os.Getenv("PORT"),
		UpstreamHost:    "api.groq.com",
		UseTLS:          true,
		RewriteModelsFn: proxy.RewriteAllModelsWithUsage("llm"),
		PathPrefix:      "/openai",
		Name:            "Groq",
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
