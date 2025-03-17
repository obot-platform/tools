package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/obot-platform/tools/search/google/googlecustomsearch/pkg/googlecustomsearchengine"
)

func exitError(err error) {
	if err != nil {
		fmt.Printf("google custom search failed: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	ctx := context.Background()
	res, err := search(ctx)
	exitError(err)
	fmt.Println(res)
}

func search(ctx context.Context) (string, error) {
	query := os.Getenv("QUERY")
	cse := googlecustomsearchengine.NewGoogleCSEFromEnv()

	results, err := cse.Search(ctx, query)
	exitError(err)

	resJSON, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	return string(resJSON), nil
}
