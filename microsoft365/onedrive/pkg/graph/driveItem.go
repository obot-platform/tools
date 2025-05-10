package graph

import (
	"context"
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func ListDriveItems(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID string, folderId string) ([]models.DriveItemable, error) {
	if driveID == "" {
		return nil, fmt.Errorf("drive ID cannot be empty")
	}

	// If no folder path is specified, list items at the root
	resp, err := client.
		Drives().
		ByDriveId(driveID).
		Items().
		ByDriveItemId(folderId).
		Children().
		Get(ctx, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to list items in drive root %s: %w", driveID, err)
	}
	return resp.GetValue(), nil
}

func ListSharedWithMeDriveItems(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) ([]models.DriveItemable, error) {
	drive, err := client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user drive: %w", err)
	}

	driveID := *drive.GetId()
	sharedWithMe, err := client.Drives().ByDriveId(driveID).SharedWithMe().GetAsSharedWithMeGetResponse(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list shared items: %w", err)
	}
	return sharedWithMe.GetValue(), nil
}

// GetDriveItem retrieves a specific item by ID from a drive
func GetDriveItem(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID string, itemID string) (models.DriveItemable, error) {
	if driveID == "" || itemID == "" {
		return nil, fmt.Errorf("drive ID and item ID cannot be empty")
	}

	item, err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(itemID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get item %s from drive %s: %w", itemID, driveID, err)
	}
	return item, nil
}

// DeleteDriveItem deletes an item from a drive
func DeleteDriveItem(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID string, itemID string) error {
	if driveID == "" || itemID == "" {
		return fmt.Errorf("drive ID and item ID cannot be empty")
	}

	err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(itemID).Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete item %s from drive %s: %w", itemID, driveID, err)
	}
	return nil
}

// DownloadDriveItem downloads the content of a drive item
func DownloadDriveItem(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID string, itemID string) ([]byte, error) {
	if driveID == "" || itemID == "" {
		return nil, fmt.Errorf("drive ID and item ID cannot be empty")
	}

	content, err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(itemID).Content().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download item %s from drive %s: %w", itemID, driveID, err)
	}
	return content, nil
}

// UploadDriveItem uploads a new file or creates a folder in a specified location
func UploadDriveItem(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID string, folderID string, name string, content []byte, isFolder bool) (models.DriveItemable, error) {
	if driveID == "" || folderID == "" || name == "" {
		return nil, fmt.Errorf("drive ID, folder ID, and name cannot be empty")
	}

	requestBody := models.NewDriveItem()
	requestBody.SetName(&name)

	// If creating a folder, set the folder facet
	if isFolder {
		requestBody.SetFolder(models.NewFolder())
		item, err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(folderID).Children().Post(ctx, requestBody, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create folder %s in drive %s: %w", name, driveID, err)
		}
		return item, nil
	}

	// For files, set the file facet
	requestBody.SetFile(models.NewFile())

	// Otherwise, create and upload file
	item, err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(folderID).Children().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create item %s in drive %s: %w", name, driveID, err)
	}

	// Upload the content
	_, err = client.Drives().ByDriveId(driveID).Items().ByDriveItemId(*item.GetId()).Content().Put(ctx, content, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upload content for item %s in drive %s: %w", name, driveID, err)
	}

	return item, nil
}

func CopyDriveItem(
	ctx context.Context,
	client *msgraphsdkgo.GraphServiceClient,
	sourceDriveID, sourceItemID,
	targetDriveID, targetFolderID, newName string,
) error {
	if sourceDriveID == "" || sourceItemID == "" || targetDriveID == "" || targetFolderID == "" {
		return fmt.Errorf("source and target Drive/Item IDs cannot be empty")
	}

	// Build the parent reference
	parentRef := models.NewItemReference()
	parentRef.SetDriveId(&targetDriveID)
	parentRef.SetId(&targetFolderID)

	// Build the copy request body
	copyBody := drives.NewItemItemsItemCopyPostRequestBody()
	copyBody.SetParentReference(parentRef)
	if newName != "" {
		copyBody.SetName(&newName)
	}

	// Perform the copy operation
	_, err := client.
		Drives().
		ByDriveId(sourceDriveID).
		Items().
		ByDriveItemId(sourceItemID).
		Copy().
		Post(ctx, copyBody, nil)
	if err != nil {
		return fmt.Errorf("failed to copy item %s from drive %s to drive %s: %w", sourceItemID, sourceDriveID, targetDriveID, err)
	}

	fmt.Printf("Successfully copied item %s from drive %s to drive %s", sourceItemID, sourceDriveID, targetDriveID)
	return nil
}

func UpdateDriveItem(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID string, itemID string, newFolderID string, newName string) (models.DriveItemable, error) {
	if driveID == "" || itemID == "" {
		return nil, fmt.Errorf("drive ID and item ID cannot be empty")
	}

	requestBody := models.NewDriveItem()
	if newFolderID != "" {
		parentReference := models.NewItemReference()
		parentReference.SetId(&newFolderID)
		requestBody.SetParentReference(parentReference)
	}

	if newName != "" {
		requestBody.SetName(&newName)
	}

	items, err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(itemID).Patch(context.Background(), requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update item %s in drive %s: %w", itemID, driveID, err)
	}
	return items, nil
}
