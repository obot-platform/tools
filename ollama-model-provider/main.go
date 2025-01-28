package main

import (
	"os"
	"strings"

	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

func cleanHost(host string) string {
	return strings.TrimRight(host, "/")
}

func main() {
	host := os.Getenv("OBOT_OLLAMA_MODEL_PROVIDER_HOST")
	if host == "" {
		host = "127.0.0.1:11434"
	}
	host = cleanHost(host)

	cfg := &proxy.Config{
		APIKey:          "",
		ListenPort:      os.Getenv("PORT"),
		BaseURL:         "http://" + host + "/v1",
		RewriteModelsFn: proxy.RewriteAllModelsWithUsage("llm"),
		Name:            "Ollama",
	}

	if len(os.Args) > 1 && os.Args[1] == "validate" {
		if err := cfg.Validate("/tools/ollama-model-provider/validate"); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := proxy.Run(cfg); err != nil {
		panic(err)
	}
}
