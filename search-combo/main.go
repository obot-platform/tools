package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gptscript-ai/go-gptscript"
)

func envOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func exitError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type searchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"` // likely just a snippet, depending on the search engine
}

type searchEngine struct {
	toolRef      string
	opts         func(query string) gptscript.Options
	resultMapper func(map[string]any) searchResult
}

type scrapeFunc func(ctx context.Context, c *gptscript.GPTScript, urls []string) (string, error)

func scrapeFuncSingleUrl(toolRef string, opts func(url string) gptscript.Options) scrapeFunc {
	return func(ctx context.Context, c *gptscript.GPTScript, urls []string) (string, error) {
		txts := make([]string, len(urls))
		for i, r := range urls {
			o := opts(r)
			slog.Info("scraping", "url", r)
			resp, err := c.Run(ctx, toolRef, o)
			if err != nil {
				return "", err
			}
			defer resp.Close()
			t, err := resp.Text()
			if err != nil {
				return "", err
			}
			txts[i] = t
		}
		return strings.Join(txts, "\n"), nil
	}
}

var (
	searchEngines = map[string]searchEngine{
		"googlecustomsearch": searchEngine{
			toolRef: "github.com/gptscript-ai/tools/search/google/googlecustomsearch",
			opts: func(query string) gptscript.Options {
				return gptscript.Options{
					Input: fmt.Sprintf(`{ "query": "%s" }`, query),
				}
			},
			resultMapper: func(r map[string]any) searchResult {
				sr := searchResult{}
				if t, ok := r["Title"]; ok && t != nil {
					sr.Title = t.(string)
				}
				if l, ok := r["Link"]; ok && l != nil {
					sr.URL = l.(string)
				}
				if s, ok := r["Snippet"]; ok && s != nil {
					sr.Content = s.(string)
				}
				return sr
			},
		},
	}
	scrapers = map[string]scrapeFunc{
		"colly": scrapeFuncSingleUrl("github.com/gptscript-ai/tools/colly-scraper", func(url string) gptscript.Options {
			return gptscript.Options{
				Input: fmt.Sprintf(`{"url": "%s"}`, url),
			}
		}),
		"firecrawl": scrapeFuncSingleUrl("Scrape URL from github.com/gptscript-ai/tools/firecrawl", func(url string) gptscript.Options {
			return gptscript.Options{
				Input: fmt.Sprintf(`{"url": "%s"}`, url),
			}
		}),
	}
	scrapeModes = map[string]func(ctx context.Context, c *gptscript.GPTScript, query string, results []searchResult) ([]string, error){
		"all": func(ctx context.Context, c *gptscript.GPTScript, query string, results []searchResult) ([]string, error) {
			urls := make([]string, len(results))
			for i, r := range results {
				urls[i] = r.URL
			}
			return urls, nil
		},
		"llm": func(ctx context.Context, c *gptscript.GPTScript, query string, results []searchResult) ([]string, error) {
			resJson, err := json.Marshal(results)
			if err != nil {
				return nil, err
			}
			resp, err := c.Evaluate(ctx, gptscript.Options{}, gptscript.ToolDef{
				Name: "reduce",
				Instructions: fmt.Sprintf(`
You are a web search advisor, perfecting the way we search the internet.
You are given a list of web sources with a title, a text snippet and a link.
For the given search query, please select the most relevant sources to scrape.
Return only the URLs of the selected relevant sources, separated by commas. Do not add any other text or formatting.
SEARCH_QUERY: %s
SEARCH_RESULTS: 
%s
`, query, resJson),
			})
			if err != nil {
				return nil, err
			}
			defer resp.Close()
			t, err := resp.Text()
			if err != nil {
				return nil, err
			}
			var urls []string
			u := strings.Split(t, ",")
			for _, i := range u {
				url := strings.TrimSpace(i)
				if url != "" {
					urls = append(urls, url)
				}
			}
			return urls, nil
		},
	}
	answerModes = map[string]func(ctx context.Context, c *gptscript.GPTScript, query, txt string) (string, error){
		"all": func(_ context.Context, _ *gptscript.GPTScript, _, txt string) (string, error) {
			return txt, nil
		},
		"llm": func(ctx context.Context, c *gptscript.GPTScript, query, txt string) (string, error) {
			resp, err := c.Evaluate(ctx, gptscript.Options{}, gptscript.ToolDef{
				Name: "answer",
				Instructions: fmt.Sprintf(`
You are a search engine answer generator.
Based on the given website texts, generate an answer to the search query.
Answer in the following JSON format. Do not add markdown highlighting or any other text around it: {"answer": "<your answer>", "sources": ["<source-link-1>", "<source-link-2>"]}
SEARCH_QUERY: %s
WEBSITE_TEXTS: 
%s
`, query, txt),
			})
			if err != nil {
				return "", err
			}
			defer resp.Close()
			return resp.Text()
		},
	}
)

func main() {
	// Config
	search := envOrDefault("SEARCH", "googlecustomsearch")
	scraper := envOrDefault("SCRAPER", "colly")
	scrapeMode := envOrDefault("SCRAPEMODE", "all")
	answerMode := envOrDefault("ANSWERMODE", "all")
	query := envOrDefault("QUERY", "")

	if query == "" {
		exitError(fmt.Errorf("no search query"))
	}

	var ok bool
	var searchTool searchEngine
	var scraperTool scrapeFunc
	var scrapeModeFunc func(context.Context, *gptscript.GPTScript, string, []searchResult) ([]string, error)
	var answerFunc func(context.Context, *gptscript.GPTScript, string, string) (string, error)
	if searchTool, ok = searchEngines[search]; !ok {
		exitError(fmt.Errorf("unknown search engine: %s", search))
	}
	if scrapeModeFunc, ok = scrapeModes[scrapeMode]; !ok {
		exitError(fmt.Errorf("unknown scrape mode: %s", scrapeMode))
	}
	if scraperTool, ok = scrapers[scraper]; !ok {
		exitError(fmt.Errorf("unknown scraper: %s", scraper))
	}
	if answerFunc, ok = answerModes[answerMode]; !ok {
		exitError(fmt.Errorf("unknown answer mode: %s", answerMode))
	}

	// Init
	g, err := gptscript.NewGPTScript()
	exitError(err)

	ctx := context.Background()

	// Search
	slog.Info("searching", "query", query)
	resp, err := g.Run(ctx, searchTool.toolRef, searchTool.opts(query))
	exitError(err)
	defer resp.Close()
	var r []map[string]any
	t, err := resp.Bytes()
	exitError(err)
	err = json.Unmarshal(t, &r)
	exitError(err)
	var searchResults []searchResult
	for _, v := range r {
		if v != nil {
			searchResults = append(searchResults, searchTool.resultMapper(v))
		}
	}

	// (Optional) narrow down search results if configured
	slog.Info("narrowing down search", "urls", len(searchResults))
	urls, err := scrapeModeFunc(ctx, g, query, searchResults)
	exitError(err)

	// (Optional) scrape search results
	slog.Info("scraping", "urls", len(urls))
	txt, err := scraperTool(ctx, g, urls)
	exitError(err)

	// answer
	slog.Info("answering", "query", query, "answerMode", answerMode)
	answer, err := answerFunc(ctx, g, query, txt)
	exitError(err)

	fmt.Println(answer)
}
