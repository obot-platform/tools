package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/obot-platform/tools/website-cleaner/pkg/clean"
	"github.com/sirupsen/logrus"
)

func main() {
	input := os.Getenv("INPUT")
	output := os.Getenv("OUTPUT")

	logOut := logrus.New()
	logOut.SetOutput(os.Stdout)
	logOut.SetFormatter(&logrus.JSONFormatter{})
	logErr := logrus.New()
	logErr.SetOutput(os.Stderr)

	ctx := context.Background()
	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to create gptscript client, error: %v", err)).Error()
		os.Exit(0)
	}

	if input == "" {
		logOut.WithError(fmt.Errorf("input is empty")).Error()
		os.Exit(0)
	}

	if output == "" {
		logOut.WithError(fmt.Errorf("output is empty")).Error()
		os.Exit(0)
	}

	inputFile, err := gptscriptClient.ReadFileInWorkspace(ctx, input)
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to read input file %q: %v", input, err)).Error()
		os.Exit(0)
	}

	originalSize := len(inputFile)

	// Clean HTML programmatically
	cleaned, err := clean.Clean(inputFile)
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to clean html: %v", err)).Error()
		os.Exit(0)
	}

	// transform to Markdown
	markdown, err := clean.ToMarkdown(cleaned)
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to convert html to markdown: %v", err)).Error()
		os.Exit(0)
	}
	markdownSize := len(markdown)
	logErr.Infof("[%s] Original HTML size: %d, Converted Markdown size: %d", input, originalSize, markdownSize)

	if err := gptscriptClient.WriteFileInWorkspace(ctx, output, []byte(markdown)); err != nil {
		logOut.WithError(fmt.Errorf("failed to write output file %q: %v", output, err)).Error()
		os.Exit(0)
	}

	logErr.Infof("Output written to %s", output)
}
