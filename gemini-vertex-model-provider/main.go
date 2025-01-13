package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/obot-platform/tools/gemini-vertex-model-provider/server"
	"golang.org/x/oauth2/google"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()

	args := os.Args[1:]
	if len(args) == 1 && args[0] == "validate" {
		if err := validate(ctx); err != nil {
			fmt.Printf("{\"error\": \"%s\"}\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	c, err := configure(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	if err := server.Run(c, port); err != nil {
		panic(err)
	}
}

func validate(ctx context.Context) error {
	_, err := configure(ctx)
	return err
}

func configure(ctx context.Context) (*genai.Client, error) {
	// Ensure that we have some valid credentials JSON data
	credsJSON := os.Getenv("OBOT_GEMINI_VERTEX_MODEL_PROVIDER_GOOGLE_CREDENTIALS_JSON")
	if credsJSON == "" {
		return nil, fmt.Errorf("google application credentials content is required")
	}

	var creds map[string]any
	if err := json.Unmarshal([]byte(credsJSON), &creds); err != nil {
		return nil, fmt.Errorf("failed to parse google application credentials json: %w", err)
	}

	gcreds, err := google.CredentialsFromJSON(ctx, []byte(credsJSON), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to parse google credentials JSON: %w", err)
	}

	// Ensure that we have a Project ID set
	var pid string
	if p, ok := creds["project_id"]; ok {
		pid = p.(string)
	} else {
		pid = os.Getenv("OBOT_GEMINI_VERTEX_MODEL_PROVIDER_GOOGLE_CLOUD_PROJECT")
	}
	if pid == "" {
		return nil, fmt.Errorf("google cloud project id is required")
	}

	// Ensure that we have a Location set
	var loc string
	if l, ok := creds["location"]; ok {
		loc = l.(string)
	} else {
		loc = os.Getenv("OBOT_GEMINI_VERTEX_MODEL_PROVIDER_GOOGLE_CLOUD_LOCATION")
	}
	if loc == "" {
		return nil, fmt.Errorf("google cloud location is required")
	}

	cc := &genai.ClientConfig{
		Backend:     genai.BackendVertexAI,
		Credentials: gcreds,
		Project:     pid,
		Location:    loc,
	}

	client, err := genai.NewClient(ctx, cc)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return client, nil
}
