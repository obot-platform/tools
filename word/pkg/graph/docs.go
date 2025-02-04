package graph

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"code.sajari.com/docconv/v2"
	kiota "github.com/microsoft/kiota-abstractions-go"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
)

type DocInfo struct {
	ID, Name string
}

func (d DocInfo) String() string {
	return fmt.Sprintf("Name: %s\nID: %s", d.Name, d.ID)
}

// getItemByPath retrieves a drive item (file or folder) by its path relative to the drive root.
func getItemByPath(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID, path string) (graphmodels.DriveItemable, error) {
	// Build the URL using the Graph endpoint:
	// GET /drives/{drive-id}/root:/{item-path}
	requestInfo := kiota.NewRequestInformation()
	requestInfo.UrlTemplate = "{+baseurl}/drives/{driveid}/root:/{itempath}"
	// Note: URL-encode the path as needed.
	requestInfo.PathParameters = map[string]string{
		"baseurl": client.RequestAdapter.GetBaseUrl(),
	}
	requestInfo.PathParametersAny = map[string]any{
		"driveid":  driveID,
		"itempath": path,
	}
	requestInfo.Method = kiota.GET

	u, err := requestInfo.GetUri()
	if err != nil {
		return nil, fmt.Errorf("failed to get URI: %w", err)
	}
	slog.Info("Getting item by path", "path", path, "drive", driveID, "url", u)

	// Use the factory function to create a new DriveItem instance.
	res, err := client.RequestAdapter.Send(ctx, requestInfo, graphmodels.CreateDriveItemFromDiscriminatorValue, nil)
	if err != nil {
		if strings.HasSuffix(err.Error(), "404") {
			return nil, fmt.Errorf("item not found: %w", err)
		}
		return nil, err
	}

	driveItem, ok := res.(graphmodels.DriveItemable)
	if !ok {
		return nil, fmt.Errorf("unexpected response type for uploaded drive item")
	}
	return driveItem, nil
}

// ensureFolderExists walks the folder path (which may include nested folders)
// and creates any folder that does not exist. It returns the DriveItem for the final folder.
func ensureFolderExists(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID, folderPath string) (graphmodels.DriveItemable, error) {
	// Normalize and split the folder path (e.g. "FolderA/FolderB").
	parts := strings.Split(strings.Trim(folderPath, "/"), "/")
	// Start at the drive root.
	currentFolderID := "root"
	var currentItem graphmodels.DriveItemable

	// Build the path progressively.
	for idx, part := range parts {
		// Build the relative path from the root up to the current folder.
		currentPath := strings.Join(parts[:idx+1], "/")
		// Try to get the folder by path.
		item, err := getItemByPath(ctx, client, driveID, currentPath)
		if err != nil {
			if !strings.Contains(err.Error(), "item not found") {
				return nil, fmt.Errorf("failed to get item by path %q: %w", currentPath, err)
			}
			// Assume an error indicates the folder was not found.
			// Create the folder in the current parent folder.
			newFolder := graphmodels.NewDriveItem()
			newFolder.SetName(&part)
			// Mark the item as a folder.
			newFolder.SetFolder(graphmodels.NewFolder())
			// Set conflict behavior to "rename" (or "fail") to avoid naming conflicts.
			newFolder.SetAdditionalData(map[string]any{
				"@microsoft.graph.conflictBehavior": "fail",
			})
			createdFolder, err := client.Drives().
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

// uploadFileContent uploads file content as a new drive item under the specified parent folder.
func uploadFileContent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID, parentID, filename, content string) (graphmodels.DriveItemable, error) {
	if parentID == "" {
		parentID = "root"
	}

	// Check if file exists
	doc, err := getItemByPath(ctx, client, driveID, filename)
	if err != nil {
		if !strings.Contains(err.Error(), "item not found") {
			return nil, fmt.Errorf("failed to get item by path %q: %w", parentID+"/"+filename, err)
		}
	}

	// Build the URL for a simple upload:
	// PUT /drives/{drive-id}/items/{parent-id}:/{filename}:/content
	requestInfo := kiota.NewRequestInformation()
	requestInfo.PathParameters = map[string]string{
		"baseurl": client.RequestAdapter.GetBaseUrl(), // for some weird reason, this is deprecated, but also the only way to get it working
	}
	requestInfo.Method = kiota.PUT
	requestInfo.SetStreamContentAndContentType([]byte(content), "text/plain")

	if doc == nil {
		slog.Info("File does not exist. Creating.", "name", filename)
		requestInfo.UrlTemplate = "{+baseurl}/drives/{driveid}/items/{parentid}:/{filename}:/content"
		requestInfo.PathParametersAny = map[string]any{
			"driveid":  driveID,
			"parentid": parentID,
			// URL-encode the filename if necessary.
			"filename": filename,
		}
	} else {
		slog.Info("File exists. Updating.", "name", filename)
		requestInfo.UrlTemplate = "{+baseurl}/drives/{driveid}/items/{itemid}/content"
		requestInfo.PathParametersAny = map[string]any{
			"driveid": driveID,
			"itemid":  deref(doc.GetId()),
		}
	}
	u, err := requestInfo.GetUri()
	if err != nil {
		return nil, fmt.Errorf("failed to get URI: %w", err)
	}
	slog.Info("Uploading file", "name", filename, "parent", parentID, "drive", driveID, "url", u)

	errorMapping := kiota.ErrorMappings{
		"XXX": odataerrors.CreateODataErrorFromDiscriminatorValue,
	}

	res, err := client.RequestAdapter.Send(ctx, requestInfo, graphmodels.CreateDriveItemFromDiscriminatorValue, errorMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	driveItem, ok := res.(graphmodels.DriveItemable)
	if !ok {
		return nil, fmt.Errorf("unexpected response type for uploaded drive item")
	}
	return driveItem, nil
}

// CreateDoc creates (or uploads) a new document with the given name and content
// into the specified directory (dir) in the user's OneDrive.
func CreateDoc(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, dir, name, content string) (string, string, error) {
	// Get the user's drive.
	drive, err := client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to get drive: %w", err)
	}
	driveID := deref(drive.GetId())

	// Ensure the target folder exists.
	folderID := "root"
	if dir != "" {
		folderItem, err := ensureFolderExists(ctx, client, driveID, dir)
		if err != nil {
			return "", "", fmt.Errorf("failed to ensure folder exists: %w", err)
		}
		folderID = deref(folderItem.GetId())
	}

	// Upload the file into the folder.
	uploadedItem, err := uploadFileContent(ctx, client, driveID, folderID, name, content)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file: %w", err)
	}
	if uploadedItem == nil {
		return "", "", fmt.Errorf("failed to upload file: uploaded item is nil")
	}
	slog.Info("Uploaded file", "name", name, "id", deref(uploadedItem.GetId()))
	return name, deref(uploadedItem.GetId()), nil
}

func ListDocs(ctx context.Context, c *msgraphsdkgo.GraphServiceClient) ([]DocInfo, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	opts := &drives.ItemSearchWithQRequestBuilderGetRequestConfiguration{
		QueryParameters: &drives.ItemSearchWithQRequestBuilderGetQueryParameters{
			// Request that these fields are returned in the response.
			Select: []string{"id", "name", "parentReference"},
		},
		// You can also set headers or options if needed.
	}
	docs, err := c.Drives().
		ByDriveId(deref(drive.GetId())).
		SearchWithQ(ptr("docx")).
		GetAsSearchWithQGetResponse(ctx, opts)
	if err != nil {
		return nil, err
	}

	var infos []DocInfo
	for _, info := range docs.GetValue() {
		infos = append(infos, DocInfo{
			ID:   deref(info.GetId()),
			Name: deref(info.GetName()),
		})
	}

	return infos, nil
}

func GetDocByPath(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, path string) (string, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	doc, err := getItemByPath(ctx, c, deref(drive.GetId()), path)
	if err != nil {
		if strings.Contains(err.Error(), "item not found") {
			return "", fmt.Errorf("doc not found")
		}
		return "", err
	}

	parentPath := deref(doc.GetParentReference().GetPath())
	slog.Info("Getting doc by path", "path", path, "parentPath", parentPath)

	docContent, err := c.Drives().ByDriveId(deref(drive.GetId())).Items().ByDriveItemId(deref(doc.GetId())).Content().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	content, err := docconv.Convert(bytes.NewReader(docContent), "application/vnd.ms-word", true)
	if err != nil {
		return "", fmt.Errorf("failed to convert doc: %w", err)
	}

	return content.Body, nil
}

func GetDoc(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, docID string) (string, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	doc, err := c.Drives().ByDriveId(deref(drive.GetId())).Items().ByDriveItemId(docID).Content().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	content, err := docconv.Convert(bytes.NewReader(doc), "application/vnd.ms-word", true)
	if err != nil {
		return "", err
	}

	return content.Body, nil
}

func ptr[T any](v T) *T {
	return &v
}

func deref[T any](v *T) (r T) {
	if v != nil {
		return *v
	}
	return
}
