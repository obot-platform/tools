package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gptscript-ai/go-gptscript"
)

type output struct {
	Results []subqueryResults `json:"subqueryResults"`
}

type subqueryResults struct {
	Subquery        string     `json:"subquery"`
	ResultDocuments []document `json:"resultDocuments"`
}

type document struct {
	ID       string   `json:"id"`
	Content  string   `json:"content,omitempty"`
	Metadata metadata `json:"metadata,omitempty"`
}

type metadata struct {
	Source            string `json:"source,omitempty"`
	WorkspaceID       string `json:"workspaceID,omitempty"`
	URL               string `json:"url,omitempty"`
	Pages             string `json:"pages,omitempty"`
	Page              int    `json:"page,omitempty"`
	TotalPages        int    `json:"totalPages,omitempty"`
	FileSize          int    `json:"fileSize,omitempty"`
	WorkspaceFileName string `json:"workspaceFileName,omitempty"` // workspaceFileName is the location of the converted file, not the original file - e.g. <path>/foo.pdf.json
}

type hit struct {
	URL      string `json:"url,omitempty"`      // URL should be the original source of the document (Web URL, OneDrive Link, Workspace File, etc.)
	Location string `json:"location,omitempty"` // Location should be the location of the result in the original source (page numbers, etc.)
	Content  string `json:"content,omitempty"`  // Content should be the text content of the document
}

type inputContent struct {
	Documents []document `json:"documents"`
}

func main() {
	var (
		output            output
		out               = gptscript.GetEnv("OUTPUT", "")
		client, clientErr = gptscript.NewGPTScript()
		ctx               = context.Background()
	)

	// This is ugly code, I know. Beauty comes later. Cleaned up a little. Still room for improvement.

	if clientErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create gptscript client: %v\n", clientErr)
	}

	if err := json.Unmarshal([]byte(out), &output); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to unmarshal output: %v\n", err)
		fmt.Print(out)
		return
	}

	var (
		outDocs      []hit
		wg           sync.WaitGroup
		fullyFetched = map[string]int{} // fullyFetched is a map of files that have been fully fetched from the workspace - the value is the index in outDocs
		budget       = 120_000
	)

	for _, result := range output.Results {
		if len(outDocs) >= 10 {
			break
		}
		for _, doc := range result.ResultDocuments {
			filename := doc.Metadata.WorkspaceFileName

			// We parse the location regardless of the file potentially being fully fetched already to preserve the
			// source reference metadata (i.e. where in the document the information was found).
			// This is a UX thing to help users with manual proofreading of answers.
			var location string
			if doc.Metadata.Pages != "" {
				location = "Pages " + doc.Metadata.Pages
			} else if doc.Metadata.Page > 0 {
				location = fmt.Sprintf("Page %d", doc.Metadata.Page)
			}
			if location != "" && doc.Metadata.TotalPages > 0 {
				location = fmt.Sprintf("%s of %d", location, doc.Metadata.TotalPages)
				_, _ = fmt.Fprintf(os.Stderr, "result doc in file %q at %q\n", filename, location)
			}

			if ffi, ok := fullyFetched[filename]; ok {
				if location != "" {
					outDocs[ffi].Location += " and " + location
				}
				continue
			}

			// url should be the original source of the document (Web URL, OneDrive Link, Workspace File, etc.)
			var url string
			if strings.HasPrefix(doc.Metadata.Source, "ws://") {
				url = doc.Metadata.Source
			} else {
				url = doc.Metadata.URL
			}
			_, _ = fmt.Fprintf(os.Stderr, "result doc url %q\n", url)

			outDocs = append(outDocs, hit{
				URL:      url,
				Content:  doc.Content,
				Location: location,
			})

			index := len(outDocs) - 1

			if index < 3 && clientErr == nil {
				fileSize := doc.Metadata.FileSize
				workspaceID := doc.Metadata.WorkspaceID
				if fileSize > 5_000 && fileSize < budget && workspaceID != "" {
					_, _ = fmt.Fprintf(os.Stderr, "fetching full file %q from workspace: %d bytes\n", filename, fileSize)
					fullyFetched[filename] = index
					budget -= fileSize
					wg.Add(1)

					go func() {
						defer wg.Done()

						content, err := client.ReadFileInWorkspace(ctx, filename, gptscript.ReadFileInWorkspaceOptions{
							WorkspaceID: workspaceID,
						})
						if err != nil {
							_, _ = fmt.Fprintf(os.Stderr, "failed to read file in workspace: %v\n", err)
							return
						}

						var sourceContent inputContent
						if err := json.Unmarshal(content, &sourceContent); err != nil {
							_, _ = fmt.Fprintf(os.Stderr, "failed to unmarshal content: %v\n", err)
							return
						}

						var buffer strings.Builder
						for _, sourceContentDocument := range sourceContent.Documents {
							buffer.WriteString(sourceContentDocument.Content)
						}

						if buffer.Len() > 0 {
							outDocs[index].Content = buffer.String()
							outDocs[index].Location = "Full Document. Specifically " + outDocs[index].Location
						}
					}()
				} else {
					_, _ = fmt.Fprintf(os.Stderr, "file %q size %d is not within range %d\n", fmt.Sprintf("%s/%s", workspaceID, filename), fileSize, budget)
				}
			}
		}
	}
	wg.Wait()
	if len(outDocs) == 0 {
		_, _ = fmt.Println("no relevant documents found")
		return
	}
	_ = json.NewEncoder(os.Stdout).Encode(outDocs)
}
