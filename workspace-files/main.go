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
	FileEnv            = os.Getenv("FILENAME")
	DirEnv             = os.Getenv("DIR")
	MaxFileSize        = 250_000
	ThreadID           = os.Getenv("OBOT_THREAD_ID")
	ProjectID          = os.Getenv("OBOT_PROJECT_ID")
	ServerURL          = os.Getenv("OBOT_SERVER_URL")
	ProjectWorkspaceID = os.Getenv("PROJECT_WORKSPACE_ID")
	ProjectScoped      = os.Getenv("PROJECT_SCOPED") == "true"
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

type fileID struct {
	Filename      string `json:"filename,omitempty"`
	ProjectScoped bool   `json:"projectScoped,omitempty"`
}

func (id fileID) String() string {
	scope := "workspace"
	if id.ProjectScoped {
		scope = "project workspace"
	}

	return fmt.Sprintf("%q %s file", id.Filename, scope)
}

func parseFileID(unparsed string) fileID {
	parsed := strings.TrimPrefix(unparsed, "project://")
	return fileID{
		Filename:      parsed,
		ProjectScoped: len(parsed) != len(unparsed),
	}
}

type data struct {
	Prompt       string            `json:"prompt,omitempty"`
	Explain      *explain          `json:"explain,omitempty"`
	Improve      *explain          `json:"improve,omitempty"`
	ChangedFiles map[string]string `json:"changedFiles,omitempty"`
}

type explain struct {
	fileID    `json:",inline"`
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

	var output strings.Builder
	if data.Explain != nil {
		fmt.Fprintf(&output, `Explain the following selection from the %s: %s\n`,
			data.Explain.fileID, inBackTicks(data.Explain.Selection))
	}

	if data.Improve != nil {
		if data.Improve.Selection == "" {
			fmt.Fprintf(&output, `Refering to the %s, %s
Write any suggested changes back to the file.`, data.Improve.fileID, data.Prompt)
		} else {
			fmt.Fprintf(&output, `Refering to the below selection from the %s, %s: %s
Write any suggested changes back to the file.`,
				data.Improve.fileID, data.Prompt, inBackTicks(data.Improve.Selection))
		}
	}

	if len(data.ChangedFiles) > 0 {
		var (
			files, projectFiles []string
			headerTemplate      = "The following files have been externally changed in the %s, re-read them if the up to date content needs to be known:"
			entryTemplate       = "%s: %s\n%s"
		)

		c, err := gptscript.NewGPTScript()
		if err == nil {
			for filename, content := range data.ChangedFiles {
				id := parseFileID(filename)

				var workspaceID string
				if id.ProjectScoped {
					if ProjectWorkspaceID == "" {
						// Project workspace isn't set up, so don't attempt to write changes to it
						continue
					}

					workspaceID = ProjectWorkspaceID
				}

				if err := c.WriteFileInWorkspace(ctx, path.Join(FilesDir, id.Filename), []byte(content), gptscript.WriteFileInWorkspaceOptions{
					WorkspaceID: workspaceID,
				}); err == nil {
					if id.ProjectScoped {
						projectFiles = append(projectFiles, fmt.Sprintf(entryTemplate, id.Filename, inBackTicks(content), "Project File"))
						continue
					}

					files = append(files, fmt.Sprintf(entryTemplate, id.Filename, inBackTicks(content), "File"))
				}

			}

			if len(projectFiles) > 0 {
				fmt.Fprintf(&output, headerTemplate, "project workspace")
				fmt.Fprint(&output, strings.Join(projectFiles, "\n"))
			}

			if len(files) > 0 {
				fmt.Fprintf(&output, headerTemplate, "workspace")
				fmt.Fprint(&output, strings.Join(files, "\n"))
			}
		}
	}

	if data.Prompt != "" {
		fmt.Fprintln(&output, data.Prompt)
	}

	fmt.Print(output.String())
}

func list(ctx context.Context, filename string) error {
	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	// List files in the thread-scoped workspace
	files, err := client.ListFilesInWorkspace(ctx, gptscript.ListFilesInWorkspaceOptions{
		Prefix: path.Join(FilesDir, filename),
	})
	if err != nil {
		return err
	}

	toPrint := make(map[string]struct{}, len(files))
	for _, file := range files {
		p := strings.TrimPrefix(file, FilesDir+"/")
		if p == "" {
			continue
		}

		parts := strings.Split(p, "/")
		if len(parts) > 1 {
			toPrint[parts[0]+"/"] = struct{}{}
			continue
		}

		toPrint[parts[0]] = struct{}{}
	}

	// List files in the project workspace if ProjectWorkspaceID is available
	var projectFiles []string
	if ProjectWorkspaceID != "" {
		projectFiles, err = client.ListFilesInWorkspace(ctx, gptscript.ListFilesInWorkspaceOptions{
			Prefix:      path.Join(FilesDir, filename),
			WorkspaceID: ProjectWorkspaceID,
		})
		if err != nil {
			return err
		}
	}

	projectToPrint := make(map[string]struct{}, len(projectFiles))
	for _, file := range projectFiles {
		p := strings.TrimPrefix(file, FilesDir+"/")
		if p == "" {
			continue
		}

		parts := strings.Split(p, "/")
		if len(parts) > 1 {
			projectToPrint[parts[0]+"/"] = struct{}{}
			continue
		}

		projectToPrint[parts[0]] = struct{}{}
	}

	// Output with XML-like tags using string builder
	var output strings.Builder
	output.WriteString("<files_in_workspace>\n")
	if len(toPrint) > 0 {
		output.WriteString("List of files currently in workspace:\n")
		output.WriteString(strings.Join(slices.Sorted(maps.Keys(toPrint)), "\n"))
		output.WriteString("\n")
	} else {
		output.WriteString("No files found in workspace\n")
	}
	output.WriteString("</files_in_workspace>\n\n")

	output.WriteString("<files_in_project_workspace>\n")
	if len(projectToPrint) > 0 {
		output.WriteString("List of files currently in project workspace:\n")
		output.WriteString(strings.Join(slices.Sorted(maps.Keys(projectToPrint)), "\n"))
		output.WriteString("\n")
	} else {
		output.WriteString("No files found in project workspace\n")
	}
	output.WriteString("</files_in_project_workspace>\n")

	fmt.Print(output.String())

	return nil
}

func readNonPlainOrLargeFile(ctx context.Context, filename, workspaceID string) (string, error) {
	// forward to a tool to handle non-plain text files or large files
	client, err := gptscript.NewGPTScript()
	if err != nil {
		return "", err
	}

	newData := map[string]string{"input_file": filename}
	jsonData, err := json.Marshal(newData)
	if err != nil {
		return "", err
	}

	var text string
	run, err := client.Run(ctx, "github.com/obot-platform/tools/file-summarizer/tool.gpt", gptscript.Options{
		Input:     string(jsonData),
		Workspace: workspaceID,
	})
	if err != nil {
		return "", err
	}

	text, err = run.Text()
	if err != nil {
		return "", err
	}

	return text, nil
}

func read(ctx context.Context, filename string) error {
	// Attempt to get the workspace ID.
	var workspaceID string
	if ProjectScoped {
		if ProjectWorkspaceID == "" {
			return fmt.Errorf("PROJECT_WORKSPACE_ID is not set")
		}

		workspaceID = ProjectWorkspaceID
	}

	// Check if the file extension is not plain text. If it is, forward it the a separate tool to handle it.
	var triedNonPlain bool
	for _, ext := range nonPlainTextFileTypes {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			text, err := readNonPlainOrLargeFile(ctx, filename, workspaceID) // hand it to the tool to ingest it and potentially summarize it
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

	data, err := client.ReadFileInWorkspace(ctx, path.Join(FilesDir, filename), gptscript.ReadFileInWorkspaceOptions{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return err
	}

	if len(data) > MaxFileSize {
		if triedNonPlain {
			return fmt.Errorf("file size exceeds %d bytes", MaxFileSize)
		}

		text, err := readNonPlainOrLargeFile(ctx, filename, workspaceID) // hand it to the tool to summarize it
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
	// Attempt to get the workspace ID
	var workspaceID string
	if ProjectScoped {
		if ProjectWorkspaceID == "" {
			return fmt.Errorf("PROJECT_WORKSPACE_ID is not set")
		}

		workspaceID = ProjectWorkspaceID
	}

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

	return client.WriteFileInWorkspace(ctx, path.Join(FilesDir, filename), []byte(content), gptscript.WriteFileInWorkspaceOptions{
		WorkspaceID: workspaceID,
	})
}

func copy(ctx context.Context, filename, toFilename string) error {
	// Attempt to get the workspace ID
	var workspaceID string
	if ProjectScoped {
		if ProjectWorkspaceID == "" {
			return fmt.Errorf("PROJECT_WORKSPACE_ID is not set")
		}

		workspaceID = ProjectWorkspaceID
	}

	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	data, err := client.ReadFileInWorkspace(ctx, path.Join(FilesDir, filename), gptscript.ReadFileInWorkspaceOptions{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return err
	}

	return client.WriteFileInWorkspace(ctx, path.Join(FilesDir, toFilename), data, gptscript.WriteFileInWorkspaceOptions{
		WorkspaceID: workspaceID,
	})
}

func downloadURL(ctx context.Context, filename string) error {
	// Attempt to get the workspace ID and thread ID for the requested workspace scope
	var (
		workspaceID string
		threadID    string
	)
	if ProjectScoped {
		if ProjectWorkspaceID == "" || ProjectID == "" {
			return fmt.Errorf("PROJECT_WORKSPACE_ID and OBOT_PROJECT_ID environment variables are required")

		}

		workspaceID = ProjectWorkspaceID
		threadID = ProjectID
	} else {
		if ThreadID == "" {
			return fmt.Errorf("OBOT_THREAD_ID environment variable is required")
		}

		threadID = ThreadID
	}

	if ServerURL == "" {
		return fmt.Errorf("OBOT_SERVER_URL environment variable is required")
	}

	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	// Check if file exists
	_, err = client.StatFileInWorkspace(ctx, path.Join(FilesDir, filename), gptscript.StatFileInWorkspaceOptions{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("%s/api/threads/%s/file/%s\n", ServerURL, threadID, filename)

	return nil
}
