package main

import (
	"fmt"
	"os"

	"github.com/obot-platform/tools/deepseek-model-provider/server"
	"github.com/obot-platform/tools/deepseek-model-provider/validate"
)

func main() {
	apiKey := os.Getenv("OBOT_DEEPSEEK_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("OBOT_DEEPSEEK_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 1 && args[0] == "validate" {
		if err := validate.Run(apiKey); err != nil {
			fmt.Printf("{\"error\": \"%s\"}\n", err)
		}
		os.Exit(0)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	if err := server.Run(apiKey, port); err != nil {
		panic(err)
	}
}
