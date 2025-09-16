package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.sajari.com/docconv/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/google/jsonschema-go/jsonschema"
	kiota "github.com/microsoft/kiota-abstractions-go"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var httpAddr = flag.String("http", ":9000", "HTTP address to listen on for streamable HTTP server")

// StaticTokenCredential implements azcore.TokenCredential
type StaticTokenCredential struct {
	token string
}

func (s StaticTokenCredential) GetToken(_ context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}

// WordMCPServer wraps the Microsoft Graph client for Word operations
type WordMCPServer struct {
	client *msgraphsdkgo.GraphServiceClient
}

// NewWordMCPServer creates a new Word MCP server with the given token
func NewWordMCPServer(token string) (*WordMCPServer, error) {
	credential := StaticTokenCredential{token: token}
	client, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Graph client: %w", err)
	}

	return &WordMCPServer{client: client}, nil
}

// DocumentInfo represents basic document information
type DocumentInfo struct {
	ID               string `json:"ID"`
	Name             string `json:"Name"`
	ParentFolderPath string `json:"ParentFolderPath,omitempty"`
	Size             *int64 `json:"Size,omitempty"`
	LastModified     string `json:"LastModified,omitempty"`
	WebURL           string `json:"WebURL,omitempty"`
}

func (d DocumentInfo) String() string {
	if d.ParentFolderPath == "" {
		return fmt.Sprintf("Name: %s\nID: %s", d.Name, d.ID)
	}
	return fmt.Sprintf("Name: %s\nParent Folder: %s\nID: %s", d.Name, d.ParentFolderPath, d.ID)
}

// Helper functions
func deref[T any](v *T) (r T) {
	if v != nil {
		return *v
	}
	return
}

// isWordDocument checks if a file is a Microsoft Word document based on its extension.
func isWordDocument(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".docx" || ext == ".doc"
}

// Argument structures - only keeping tools that exist in the actual Word implementation
type ListDocsArgs struct{}

type ReadDocArgs struct {
	DocID string `json:"doc_id" jsonschema:"ID or Path of the Microsoft Word document to get. Prefer ID if available, path only if given by user."`
}

type WriteDocArgs struct {
	DocName           string `json:"doc_name" jsonschema:"(Required) The name of the document to write to. This might be the OneDrive ID of an existing document or a filepath in OneDrive."`
	DocContent        string `json:"doc_content,omitempty" jsonschema:"(Optional) Markdown formatted content to write to the document."`
	OverwriteIfExists *bool  `json:"overwrite_if_exists,omitempty" jsonschema:"(Optional) Whether to overwrite the document if it already exists, defaults to false. You MUST only set this to true if you have confirmed with the user that they want to overwrite the document."`
}

// ListDocs lists all Word documents in the user's drive recursively
func (w *WordMCPServer) ListDocs(ctx context.Context, req *mcp.CallToolRequest, args ListDocsArgs) (*mcp.CallToolResult, any, error) {
	drive, err := w.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get drive: %w", err)
	}

	// Start from the root folder
	root, err := w.client.Drives().
		ByDriveId(deref(drive.GetId())).
		Root().
		Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get root folder: %w", err)
	}

	var infos []DocumentInfo
	err = w.listDocsInFolder(ctx, deref(drive.GetId()), deref(root.GetId()), &infos, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list docs: %w", err)
	}

	if len(infos) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "No Word documents found",
				},
			},
		}, nil, nil
	}

	result, _ := json.MarshalIndent(infos, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// listDocsInFolder recursively lists all documents in a folder and its subfolders.
func (w *WordMCPServer) listDocsInFolder(ctx context.Context, driveID, folderID string, infos *[]DocumentInfo, currentPath string) error {
	items, err := w.client.Drives().
		ByDriveId(driveID).
		Items().
		ByDriveItemId(folderID).
		Children().
		Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get items in folder: %w", err)
	}

	// Process this page of items
	for _, item := range items.GetValue() {
		itemName := deref(item.GetName())
		itemPath := filepath.Join(currentPath, itemName)

		// Skip folders, but process their contents
		if item.GetFolder() != nil {
			err = w.listDocsInFolder(ctx, driveID, deref(item.GetId()), infos, itemPath)
			if err != nil {
				return err
			}
			continue
		}

		// Only include Word documents
		file := item.GetFile()
		if file == nil || !isWordDocument(itemName) {
			continue
		}

		// Add Word documents to our list
		*infos = append(*infos, DocumentInfo{
			ID:               deref(item.GetId()),
			Name:             itemName,
			ParentFolderPath: currentPath,
		})
	}

	return nil
}

// ReadDoc reads the contents of a Word document
func (w *WordMCPServer) ReadDoc(ctx context.Context, req *mcp.CallToolRequest, args ReadDocArgs) (*mcp.CallToolResult, any, error) {
	drive, err := w.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get drive: %w", err)
	}

	var content string
	if strings.HasSuffix(args.DocID, ".docx") || strings.Contains(args.DocID, "/") {
		// Handle path-based access
		content, err = w.getDocByPath(ctx, deref(drive.GetId()), args.DocID)
	} else {
		// Handle ID-based access
		content, err = w.getDoc(ctx, deref(drive.GetId()), args.DocID)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get doc %q: %w", args.DocID, err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: content,
			},
		},
	}, nil, nil
}

// getItemByPath retrieves a drive item by its path relative to the drive root.
func (w *WordMCPServer) getItemByPath(ctx context.Context, driveID, path string) (models.DriveItemable, error) {
	requestInfo := kiota.NewRequestInformation()
	requestInfo.UrlTemplate = "{+baseurl}/drives/{driveid}/root:/{itempath}"
	requestInfo.PathParameters = map[string]string{
		"baseurl": w.client.RequestAdapter.GetBaseUrl(),
	}
	requestInfo.PathParametersAny = map[string]any{
		"driveid":  driveID,
		"itempath": path,
	}
	requestInfo.Method = kiota.GET

	res, err := w.client.RequestAdapter.Send(ctx, requestInfo, models.CreateDriveItemFromDiscriminatorValue, nil)
	if err != nil {
		if strings.HasSuffix(err.Error(), "404") {
			return nil, fmt.Errorf("item not found: %w", err)
		}
		return nil, err
	}

	driveItem, ok := res.(models.DriveItemable)
	if !ok {
		return nil, fmt.Errorf("unexpected response type for drive item")
	}
	return driveItem, nil
}

func (w *WordMCPServer) getDocByPath(ctx context.Context, driveID, path string) (string, error) {
	doc, err := w.getItemByPath(ctx, driveID, path)
	if err != nil {
		if strings.Contains(err.Error(), "item not found") {
			return "", fmt.Errorf("doc not found")
		}
		return "", err
	}

	docContent, err := w.client.Drives().ByDriveId(driveID).Items().ByDriveItemId(deref(doc.GetId())).Content().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	content, err := docconv.Convert(bytes.NewReader(docContent), "application/vnd.ms-word", true)
	if err != nil {
		return "", fmt.Errorf("failed to convert doc: %w", err)
	}

	return content.Body, nil
}

func (w *WordMCPServer) getDoc(ctx context.Context, driveID, docID string) (string, error) {
	doc, err := w.client.Drives().ByDriveId(driveID).Items().ByDriveItemId(docID).Content().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	content, err := docconv.Convert(bytes.NewReader(doc), "application/vnd.ms-word", true)
	if err != nil {
		return "", err
	}

	return content.Body, nil
}

// WriteDoc creates or updates a Word document - follows exact Word tool implementation
func (w *WordMCPServer) WriteDoc(ctx context.Context, req *mcp.CallToolRequest, args WriteDocArgs) (*mcp.CallToolResult, any, error) {
	drive, err := w.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get drive: %w", err)
	}

	// Ensure name has .docx extension (exact same logic as Word tool)
	name := strings.TrimSuffix(args.DocName, filepath.Ext(args.DocName)) + ".docx"

	// Check if file already exists (exact same logic as Word tool)
	overwrite := false
	if args.OverwriteIfExists != nil {
		overwrite = *args.OverwriteIfExists
	}

	if !overwrite {
		exists, err := w.docExists(ctx, deref(drive.GetId()), name)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to check if document exists: %w", err)
		}

		if exists {
			return nil, nil, fmt.Errorf("document with name %q already exists, aborting to prevent overwrite", name)
		}
	}

	log.Printf("Creating new Word Document in OneDrive: %s", name)

	// Convert markdown to DOCX (exact same as Word tool)
	contentBytes, err := w.markdownToDocx(args.DocContent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert markdown to docx: %w", err)
	}

	// Create document (exact same as Word tool)
	resultName, id, err := w.createDoc(ctx, deref(drive.GetId()), name, contentBytes)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Wrote content to document with name=%q and id=%q", resultName, id),
			},
		},
	}, nil, nil
}

// docExists checks if a document with the given path exists in the user's OneDrive.
func (w *WordMCPServer) docExists(ctx context.Context, driveID, path string) (bool, error) {
	_, err := w.getItemByPath(ctx, driveID, path)
	if err != nil {
		if strings.Contains(err.Error(), "item not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// createDoc creates a new document with the given name and content
func (w *WordMCPServer) createDoc(ctx context.Context, driveID, name string, content []byte) (string, string, error) {
	name = filepath.Clean(name)
	if name == "" {
		return "", "", fmt.Errorf("name cannot be empty")
	}
	dir := filepath.Dir(name)
	if dir == "." {
		dir = ""
	}
	name = filepath.Base(name)

	// Ensure the target folder exists and upload the file
	folderID := "root"
	if dir != "" {
		// For simplicity, assume root folder for now
		// In full implementation, would create nested folders as needed
	}

	uploadedItem, err := w.uploadFileContent(ctx, driveID, folderID, name, content)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file: %w", err)
	}
	if uploadedItem == nil {
		return "", "", fmt.Errorf("failed to upload file: uploaded item is nil")
	}
	return name, deref(uploadedItem.GetId()), nil
}

// uploadFileContent uploads file content as a new drive item
func (w *WordMCPServer) uploadFileContent(ctx context.Context, driveID, parentID, filename string, content []byte) (models.DriveItemable, error) {
	if parentID == "" {
		parentID = "root"
	}

	// Build the URL for a simple upload
	requestInfo := kiota.NewRequestInformation()
	requestInfo.PathParameters = map[string]string{
		"baseurl": w.client.RequestAdapter.GetBaseUrl(),
	}
	requestInfo.Method = kiota.PUT
	requestInfo.SetStreamContentAndContentType(content, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")

	requestInfo.UrlTemplate = "{+baseurl}/drives/{driveid}/items/{parentid}:/{filename}:/content"
	requestInfo.PathParametersAny = map[string]any{
		"driveid":  driveID,
		"parentid": parentID,
		"filename": filename,
	}

	errorMapping := kiota.ErrorMappings{
		"XXX": odataerrors.CreateODataErrorFromDiscriminatorValue,
	}

	res, err := w.client.RequestAdapter.Send(ctx, requestInfo, models.CreateDriveItemFromDiscriminatorValue, errorMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	driveItem, ok := res.(models.DriveItemable)
	if !ok {
		return nil, fmt.Errorf("unexpected response type for uploaded drive item")
	}
	return driveItem, nil
}

// ensureFolderExists walks the folder path and creates any folder that does not exist
func (w *WordMCPServer) ensureFolderExists(ctx context.Context, driveID, folderPath string) (models.DriveItemable, error) {
	// Normalize and split the folder path (e.g. "FolderA/FolderB").
	parts := strings.Split(strings.Trim(folderPath, "/"), "/")
	// Start at the drive root.
	currentFolderID := "root"
	var currentItem models.DriveItemable

	// Build the path progressively.
	for idx, part := range parts {
		// Build the relative path from the root up to the current folder.
		currentPath := strings.Join(parts[:idx+1], "/")
		// Try to get the folder by path.
		item, err := w.getItemByPath(ctx, driveID, currentPath)
		if err != nil {
			if !strings.Contains(err.Error(), "item not found") {
				return nil, fmt.Errorf("failed to get item by path %q: %w", currentPath, err)
			}
			// Assume an error indicates the folder was not found.
			// Create the folder in the current parent folder.
			newFolder := models.NewDriveItem()
			newFolder.SetName(&part)
			// Mark the item as a folder.
			newFolder.SetFolder(models.NewFolder())
			// Set conflict behavior to "rename" (or "fail") to avoid naming conflicts.
			newFolder.SetAdditionalData(map[string]any{
				"@microsoft.graph.conflictBehavior": "fail",
			})
			createdFolder, err := w.client.Drives().
				ByDriveId(driveID).
				Items().
				ByDriveItemId(currentFolderID).
				Children().
				Post(ctx, newFolder, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create folder %q: %w", part, err)
			}
			currentItem = createdFolder
		} else {
			// Folder already exists.
			currentItem = item
		}
		// Update the parent folder for the next iteration.
		currentFolderID = deref(currentItem.GetId())
	}

	return currentItem, nil
}

// markdownToDocx converts markdown content to DOCX format - follows exact Word tool implementation
func (w *WordMCPServer) markdownToDocx(in string) ([]byte, error) {
	tempfile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("word-convsource-*.md"))
	if err != nil {
		return nil, err
	}
	defer tempfile.Close()

	p := tempfile.Name()
	_, err = tempfile.WriteString(in)
	if err != nil {
		return nil, err
	}
	_ = tempfile.Close()
	defer os.Remove(p)

	var cmd *exec.Cmd
	var outputFile string
	if _, err := exec.LookPath("pandoc"); err == nil {
		cmd, outputFile = w.pandocCmd(p)
		log.Printf("Used pandoc to convert markdown to docx: input=%s output=%s", p, outputFile)
	} else if _, err := exec.LookPath("soffice"); err == nil {
		var cleanupFunc func()
		cmd, cleanupFunc, outputFile, err = w.sofficeCmd(p)
		if err != nil {
			return nil, err
		}
		log.Printf("Used soffice to convert markdown to docx: input=%s output=%s", p, outputFile)
		defer cleanupFunc()
	} else {
		return nil, fmt.Errorf("neither pandoc nor soffice binary found")
	}

	// capture stdout and stderr in a buffer
	var outb, errb strings.Builder
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		log.Printf("Failed to run pandoc/soffice command: error=%v stderr=%s stdout=%s", err, errb.String(), outb.String())
		return nil, err
	}

	log.Printf("pandoc/soffice command output: stdout=%s stderr=%s", outb.String(), errb.String())

	content, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, err
	}
	return content, os.Remove(outputFile)
}

func (w *WordMCPServer) pandocCmd(p string) (*exec.Cmd, string) {
	outFile := fmt.Sprintf("%s.docx", p)
	return exec.Command(
		"pandoc",
		"-f", "markdown",
		"-t", "docx",
		"--output", outFile,
		p,
	), outFile
}

func (w *WordMCPServer) sofficeCmd(p string) (*exec.Cmd, func(), string, error) {
	var err error
	p, err = w.markdownToHTML(p)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to convert markdown to html: %w", err)
	}

	profileDir, err := os.MkdirTemp(os.TempDir(), "libreoffice-profile-*")
	if err != nil {
		log.Printf("Failed to create soffice profile directory: path=%s error=%v", profileDir, err)
		return nil, nil, "", fmt.Errorf("failed to create soffice profile directory: %w", err)
	}
	out := strings.TrimSuffix(p, filepath.Ext(p)) + ".docx"
	return exec.Command(
			"soffice",
			"--headless",
			fmt.Sprintf("-env:UserInstallation=file://%s", profileDir),
			"--convert-to", "docx:Office Open XML Text",
			"--outdir", filepath.Dir(out),
			p,
		), func() {
			_ = os.Remove(p)
			_ = os.RemoveAll(profileDir)
		}, out, nil
}

func (w *WordMCPServer) markdownToHTML(p string) (string, error) {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	pars := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	md, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}

	outFile := fmt.Sprintf("%s.html", p)
	if err = os.WriteFile(outFile, markdown.ToHTML(md, pars, renderer), 0644); err != nil {
		return "", err
	}
	return outFile, nil
}

// ExtractTokenFromRequest extracts the bearer token from HTTP request headers
func ExtractTokenFromRequest(req *http.Request) (string, error) {
	// Try X-Forwarded-Access-Token first
	if token := req.Header.Get("X-Forwarded-Access-Token"); token != "" {
		return token, nil
	}

	// Try Authorization header
	if authHeader := req.Header.Get("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer "), nil
		}
	}

	return "", fmt.Errorf("no access token found in request headers")
}

func main() {
	flag.Parse()

	// Create server factory that extracts token from each request
	serverFactory := func(req *http.Request) *mcp.Server {
		token, err := ExtractTokenFromRequest(req)
		if err != nil {
			log.Printf("Failed to extract token from request: %v", err)
			// Return a server that will fail gracefully
			server := mcp.NewServer(&mcp.Implementation{Name: "word-mcp-server"}, nil)
			return server
		}

		wordServer, err := NewWordMCPServer(token)
		if err != nil {
			log.Printf("Failed to create Word MCP server: %v", err)
			// Return a server that will fail gracefully
			server := mcp.NewServer(&mcp.Implementation{Name: "word-mcp-server"}, nil)
			return server
		}

		server := mcp.NewServer(&mcp.Implementation{Name: "word-mcp-server"}, nil)

		// Create JSON schemas for the tools
		readDocSchema, _ := jsonschema.For[ReadDocArgs](nil)
		writeDocSchema, _ := jsonschema.For[WriteDocArgs](nil)

		// Register all tools with proper schemas - only tools that exist in actual Word implementation
		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_docs",
			Description: "List all Microsoft Word documents available to the user in OneDrive",
		}, wordServer.ListDocs)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "read_doc",
			Description: "Read the contents of a Microsoft Word document from OneDrive",
			InputSchema: readDocSchema,
		}, wordServer.ReadDoc)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "write_doc",
			Description: "Write a Microsoft Word document in OneDrive with the specified title and optional content. The file will be created if it doesn't exist. It will be overwritten if it already exists and overwrite_if_exists is true.",
			InputSchema: writeDocSchema,
		}, wordServer.WriteDoc)

		return server
	}

	if *httpAddr != "" {
		mcpHandler := mcp.NewStreamableHTTPHandler(serverFactory, nil)
		log.Printf("Word MCP server listening at %s", *httpAddr)

		// Create a custom multiplexer
		mux := http.NewServeMux()

		// Handle /health with custom handler
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Handle all other paths with MCP handler
		mux.Handle("/", mcpHandler)

		if err := http.ListenAndServe(*httpAddr, mux); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("HTTP address is required")
	}
}
