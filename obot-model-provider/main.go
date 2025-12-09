package main

import (
	"os"
	"strings"

	"github.com/obot-platform/tools/obot-model-provider/server"
)

func main() {
	obotHost := os.Getenv("OBOT_URL")
	if obotHost == "" {
		obotHost = strings.TrimPrefix(os.Getenv("OBOT_SERVER_URL"), "http://")
		if obotHost == "" {
			obotHost = "localhost:8080"
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	if err := server.Run(obotHost, port); err != nil {
		panic(err)
	}
}
