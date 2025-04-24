package main

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"path"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/gptscript-ai/go-gptscript"
)

const FilesDir = "files"

var (
	FileEnv     = os.Getenv("FILENAME")
	DirEnv      = os.Getenv("DIR")
	MaxFileSize = 250_000
	ThreadID    = os.Getenv("OBOT_THREAD_ID")
	ServerURL   = os.Getenv("OBOT_SERVER_URL")
)

var unsupportedWriteFileTypes = []string{".pdf", ".docx", ".doc", ".pptx", ".ppt", ".xlsx", ".xls", ".jpg", ".png", ".gif", ".mp3", ".mp4", ".zip", ".rar"}
var nonPlainTextFileTypes = []string{".pdf", ".pptx", ".ppt", ".docx", ".doc", ".odt", ".rtf", ".ipynb"} // ipynb is in json. we want to convert it to markdown

func main() {
	if len(os.Args) == 1 {
		fmt.Printf(`
Subcommands: read, write, copy, download-url
env: FILENAME, CONTENT, TO_FILENAME, GPTSCRIPT_WORKSPACE_DIR, OBOT_THREAD_ID, OBOT_SERVER_URL
Usage: go run main.go <path>\n`)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cmd := os.Args[1]
	if cmd == "read" && (FileEnv == "" || strings.HasSuffix(FileEnv, "/")) {
		cmd = "list"
	}

	switch cmd {
	case "input":
		input(ctx)
		return
	case "list":
		if err := list(ctx, DirEnv); err != nil {
			fmt.Printf("Failed to list %s: %v\n", FileEnv, err)
			return
		}
	case "read":
		if err := read(ctx, FileEnv); err != nil {
			fmt.Printf("Failed to read %s: %v\n", FileEnv, err)
			return
		}
	case "write":
		content := gptscript.GetEnv("CONTENT", "")
		if err := write(ctx, FileEnv, content); err != nil {
			fmt.Printf("Failed to write %s: %v\n", FileEnv, err)
			return
		}
		fmt.Printf("Wrote %d bytes\n", len(content))
	case "copy":
		toFilename := gptscript.GetEnv("TO_FILENAME", "")
		if err := copy(ctx, FileEnv, toFilename); err != nil {
			fmt.Printf("Failed to copy %s to %s: %v\n", FileEnv, toFilename, err)
			return
		}
	case "download-url":
		if err := downloadURL(ctx, FileEnv); err != nil {
			fmt.Printf("Failed to generate download URL for %s: %v\n", FileEnv, err)
			return
		}
	}
}

type data struct {
	Prompt       string            `json:"prompt,omitempty"`
	Explain      *explain          `json:"explain,omitempty"`
	Improve      *explain          `json:"improve,omitempty"`
	ChangedFiles map[string]string `json:"changedFiles,omitempty"`
}

type explain struct {
	Filename  string `json:"filename,omitempty"`
	Selection string `json:"selection,omitempty"`
}

func inBackTicks(s string) string {
	return "\n```\n" + s + "\n```\n"
}

func input(ctx context.Context) {
	var (
		input = gptscript.GetEnv("INPUT", "")
		data  data
	)

	if err := json.Unmarshal([]byte(input), &data); err != nil {
		fmt.Print(input)
		return
	}

	if data.Explain != nil {
		fmt.Printf(`Explain the following selection from the "%s" workspace file: %s`,
			data.Explain.Filename, inBackTicks(data.Explain.Selection))
	}

	if data.Improve != nil {
		if data.Improve.Selection == "" {
			fmt.Printf(`Refering to the workspace file "%s", %s
Write any suggested changes back to the file`, data.Improve.Filename, data.Prompt)
		} else {
			fmt.Printf(`Refering to the below selection from the workspace file "%s", %s: %s
Write any suggested changes back to the file.`,
				data.Improve.Filename, data.Prompt, inBackTicks(data.Improve.Selection))
		}
	}

	if len(data.ChangedFiles) > 0 {
		var printed bool
		c, err := gptscript.NewGPTScript()
		for filename, content := range data.ChangedFiles {
			if err == nil {
				if err := c.WriteFileInWorkspace(ctx, path.Join(FilesDir, filename), []byte(content)); err == nil {
					if !printed {
						printed = true
						fmt.Println("The following files have been externally changed in the workspace, re-read them if the up to date content needs to be known:")
					}
					fmt.Printf("File: %s\n%s\n", filename, inBackTicks(content))
				}
			}
		}
		fmt.Println("")
	}

	if data.Prompt != "" {
		fmt.Print(data.Prompt)
	}
}

func list(ctx context.Context, filename string) error {
	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	files, err := client.ListFilesInWorkspace(ctx, gptscript.ListFilesInWorkspaceOptions{
		Prefix: path.Join(FilesDir, filename),
	})
	if err != nil {
		return err
	}

	toPrint := map[string]struct{}{}
	for _, file := range files {
		p := strings.TrimPrefix(file, FilesDir+"/")
		if p != "" {
			parts := strings.Split(p, "/")
			if len(parts) > 1 {
				toPrint[parts[0]+"/"] = struct{}{}
			} else {
				toPrint[parts[0]] = struct{}{}
			}
		}
	}

	if len(toPrint) > 0 {
		fmt.Println(strings.Join(slices.Sorted(maps.Keys(toPrint)), "\n"))
	}

	return nil
}

func readNonPlainOrLargeFile(ctx context.Context, filename string) (string, error) {
	// forward to a tool to handle non-plain text files or large files
	client, err := gptscript.NewGPTScript()
	if err != nil {
		return "", err
	}

	var workspaceID = os.Getenv("GPTSCRIPT_WORKSPACE_ID")
	newData := map[string]string{"input_file": filename}

	jsonData, err := json.Marshal(newData)
	if err != nil {
		return "", err
	}

	run, err := client.Run(ctx, "github.com/obot-platform/tools/file-summarizer/tool.gpt", gptscript.Options{
		Input:     string(jsonData),
		Workspace: workspaceID,
	})

	text, err := run.Text()
	if err != nil {
		return "", err
	}

	return text, nil
}

func read(ctx context.Context, filename string) error {
	triedNonPlain := false
	// Check if the file extension is not plain text. If it is, forward it the a separate tool to handle it.
	for _, ext := range nonPlainTextFileTypes {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			text, err := readNonPlainOrLargeFile(ctx, filename) // hand it to the tool to ingest it and potentially summarize it
			triedNonPlain = true
			if err == nil {
				fmt.Println(string(text))
				return nil
			}
			// if failed, fallback to plain-text attempt
			break
		}
	}

	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	data, err := client.ReadFileInWorkspace(ctx, path.Join(FilesDir, filename))
	if err != nil {
		return err
	}

	if len(data) > MaxFileSize {
		if triedNonPlain {
			return fmt.Errorf("file size exceeds %d bytes", MaxFileSize)
		}

		text, err := readNonPlainOrLargeFile(ctx, filename) // hand it to the tool to summarize it
		if err != nil {
			return fmt.Errorf("file size exceeds %d bytes and failed to summarize it as plain text: %w", MaxFileSize, err)
		}
		fmt.Println(string(text))
		return nil

	}

	if utf8.Valid(data) {
		fmt.Println(string(data))
		return nil
	}

	return fmt.Errorf("file is not valid UTF-8")
}

func write(ctx context.Context, filename, content string) error {
	// Check if the file extension is not plain text. We don't support writing to non-plain text files yet.
	for _, ext := range unsupportedWriteFileTypes {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return fmt.Errorf("writing to files with extension %s is not supported", ext)
		}
	}

	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	return client.WriteFileInWorkspace(ctx, path.Join(FilesDir, filename), []byte(content))
}

func copy(ctx context.Context, filename, toFilename string) error {
	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	data, err := client.ReadFileInWorkspace(ctx, path.Join(FilesDir, filename))
	if err != nil {
		return err
	}

	return client.WriteFileInWorkspace(ctx, path.Join(FilesDir, toFilename), data)
}

func downloadURL(ctx context.Context, filename string) error {
	if ThreadID == "" || ServerURL == "" {
		return fmt.Errorf("OBOT_THREAD_ID and OBOT_SERVER_URL environment variables are required")
	}

	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	// Check if file exists
	_, err = client.StatFileInWorkspace(ctx, path.Join(FilesDir, filename))
	if err != nil {
		return err
	}

	downloadURL := fmt.Sprintf("%s/api/threads/%s/file/%s", ServerURL, ThreadID, filename)
	fmt.Println(downloadURL)
	return nil
}
