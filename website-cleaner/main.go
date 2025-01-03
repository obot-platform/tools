package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/sirupsen/logrus"

	md "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
)

var tagsToRemove = []string{
	"script, style, noscript, meta, head",
	"header", "footer", "nav", "aside", ".header", ".top", ".navbar", "#header",
	".footer", ".bottom", "#footer", ".sidebar", ".side", ".aside", "#sidebar",
	".modal", ".popup", "#modal", ".overlay", ".ad", ".ads", ".advert", "#ad",
	".lang-selector", ".language", "#language-selector", ".social", ".social-media",
	".social-links", "#social", ".menu", ".navigation", "#nav", ".breadcrumbs",
	"#breadcrumbs", "#search-form", ".search", "#search", ".share", "#share",
	".widget", "#widget", ".cookie", "#cookie",
}

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

	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(inputFile))
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to parse html input: %v", err)).Error()
		os.Exit(0)
	}

	// Clean HTML programmatically
	for _, tag := range tagsToRemove {
		doc.Find(tag).Remove()
	}

	// transform to Markdown
	converter := md.NewConverter(md.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()))
	html, err := doc.Html()
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to get html from document: %v", err)).Error()
		os.Exit(0)
	}

	sanitizedHTMLSize := len(html)

	markdown, err := converter.ConvertString(html)
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to convert html to markdown: %v", err)).Error()
		os.Exit(0)
	}

	markdownSize := len(markdown)
	logErr.Infof("[%s] Original HTML size: %d, Sanitized HTML size: %d, Converted Markdown size: %d", input, originalSize, sanitizedHTMLSize, markdownSize)

	if err := gptscriptClient.WriteFileInWorkspace(ctx, output, []byte(markdown)); err != nil {
		logOut.WithError(fmt.Errorf("failed to write output file %q: %v", output, err)).Error()
		os.Exit(0)
	}

	logErr.Infof("Output written to %s", output)

}
