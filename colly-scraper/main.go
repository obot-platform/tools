package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	cleaner "github.com/obot-platform/tools/website-cleaner/pkg/clean"
)

func exitError(err error) {
	if err != nil {
		fmt.Printf("website scraper failed: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	txt, err := scrape(os.Getenv("URL"))
	exitError(err)
	fmt.Println(txt)
}

func scrape(url string) (string, error) {
	collector := colly.NewCollector()

	doms := map[string]*goquery.Selection{}

	collector.OnHTML("body", func(e *colly.HTMLElement) {
		doms[e.Request.URL.String()] = e.DOM
	})

	err := collector.Visit(url)
	if err != nil {
		return "", err
	}

	if len(doms) == 0 {
		return "", fmt.Errorf("no body found")
	}

	txt := strings.Builder{}

	for d, b := range doms {
		txt.WriteString(fmt.Sprintf("# !!! PAGE START - URL: %s\n", d))

		// get html
		html, err := b.Html()
		if err != nil {
			return "", err
		}

		// clean html
		cleanedHTML, err := cleaner.Clean([]byte(html))
		if err != nil {
			return "", err
		}

		// convert to markdown
		markdown, err := cleaner.ToMarkdown(cleanedHTML)
		if err != nil {
			return "", err
		}

		txt.Write(markdown)
		txt.WriteString(fmt.Sprintf("\n# !!! PAGE END - URL: %s\n", d))
	}

	return txt.String(), nil
}
