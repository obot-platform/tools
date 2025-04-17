package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	apiKey := os.Getenv("VIRUS_TOTAL_API_KEY")
	if apiKey == "" {
		fmt.Println("VIRUS_TOTAL_API_KEY is not set")
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "validate" {
		if err := validate(context.Background(), apiKey); err != nil {
			fmt.Printf("{\"error\":%q}", err.Error())
			os.Exit(1)
		}
		return
	}

	port := envOrDefault("PORT", "8000")

	http.HandleFunc("/{$}", healthz(port))
	http.Handle("POST /file", &fileHandler{
		apiKey:           apiKey,
		baseURL:          strings.TrimSuffix(envOrDefault("VIRUS_TOTAL_BASE_PATH", "https://www.virustotal.com/api/v3"), "/"),
		failOnFailures:   os.Getenv("FAIL_ON_FAILURES") == "true",
		failOnSuspicious: os.Getenv("FAIL_ON_SUSPICIOUS") == "true",
		failOnTimeout:    os.Getenv("FAIL_ON_TIMEOUT") == "true",
	})

	fmt.Printf("Starting server at port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

func healthz(port string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "http://127.0.0.1:%s", port)
	}
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
