package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gptscript-ai/go-gptscript"
)

type Field struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Survey struct {
	Message string  `json:"message"`
	Fields  []Field `json:"fields"`
}

func main() {
	surveyJSON := os.Getenv("SURVEY")
	if surveyJSON == "" {
		fmt.Println("SURVEY environment variable is not set")
		os.Exit(1)
	}

	var survey Survey
	if err := json.Unmarshal([]byte(surveyJSON), &survey); err != nil {
		fmt.Printf("Failed to parse SURVEY environment variable: %v\n", err)
		os.Exit(1)
	}

	gs, err := gptscript.NewGPTScript()
	if err != nil {
		fmt.Printf("Failed to create GPTScript: %v\n", err)
		os.Exit(1)
	}

	run, err := gs.Run(context.Background(), "sys.prompt", gptscript.Options{
		Input: surveyJSON,
	})
	if err != nil {
		fmt.Printf("Failed to run GPTScript: %v\n", err)
		os.Exit(1)
	}

	text, err := run.Text()
	if err != nil {
		fmt.Printf("Failed to get result from GPTScript: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(text)
}
