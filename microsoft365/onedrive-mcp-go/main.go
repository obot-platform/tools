// Copyright 2025 The Go MCP SDK Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/google/jsonschema-go/jsonschema"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
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

// OneDriveMCPServer wraps the Microsoft Graph client for OneDrive operations
type OneDriveMCPServer struct {
	client *msgraphsdkgo.GraphServiceClient
}

// NewOneDriveMCPServer creates a new OneDrive MCP server with the given token
func NewOneDriveMCPServer(token string) (*OneDriveMCPServer, error) {
	credential := StaticTokenCredential{token: token}
	client, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Graph client: %w", err)
	}

	return &OneDriveMCPServer{client: client}, nil
}

// DriveInfo represents drive information
type DriveInfo struct {
	ID    string `json:"ID"`
	Name  string `json:"Name"`
	Type  string `json:"Type"`
	Owner string `json:"Owner"`
	URL   string `json:"URL"`
}

// DriveItemInfo represents drive item information
type DriveItemInfo struct {
	ID         string `json:"ID"`
	Name       string `json:"Name"`
	Type       string `json:"Type"`
	Size       *int64 `json:"Size,omitempty"`
	Modified   string `json:"Modified,omitempty"`
	Created    string `json:"Created,omitempty"`
	URL        string `json:"URL,omitempty"`
	ParentPath string `json:"ParentPath,omitempty"`
}

// Argument structures
type ListAllDrivesArgs struct{}

type GetDriveArgs struct {
	DriveID *string `json:"drive_id,omitempty" jsonschema:"(Optional) The ID of the drive to get details about. If unset, default to the user's personal drive."`
}

type ListDriveItemsArgs struct {
	DriveID  string  `json:"drive_id" jsonschema:"(Required) The ID of the drive to list files in."`
	FolderID *string `json:"folder_id,omitempty" jsonschema:"(Optional) The ID of the folder to list files in. If unset, default to the root folder."`
}

type ListSharedWithMeItemsArgs struct{}

type GetDriveItemArgs struct {
	DriveID string `json:"drive_id" jsonschema:"(Required) The ID of the drive to get the file from."`
	ItemID  string `json:"item_id" jsonschema:"(Required) The ID of the item to get details about."`
}

type CreateFolderArgs struct {
	FolderName string `json:"folder_name" jsonschema:"(Required) The name of the folder to create."`
	DriveID    string `json:"drive_id" jsonschema:"(Required) The ID of the drive to create the folder in."`
	FolderID   string `json:"folder_id" jsonschema:"(Required) The ID of the folder to create the folder in. Use 'root' for the root folder."`
}

type DeleteDriveItemArgs struct {
	DriveID string `json:"drive_id" jsonschema:"(Required) The ID of the drive to delete the file from."`
	ItemID  string `json:"item_id" jsonschema:"(Required) The ID of the item to delete."`
}

type CopyDriveItemArgs struct {
	SourceDriveID  string  `json:"source_drive_id" jsonschema:"(Required) The ID of the drive to copy the file from."`
	SourceItemID   string  `json:"source_item_id" jsonschema:"(Required) The ID of the item to copy."`
	TargetDriveID  string  `json:"target_drive_id" jsonschema:"(Required) The ID of the drive to copy the file to."`
	TargetFolderID string  `json:"target_folder_id" jsonschema:"(Required) The ID of the folder to copy the file to."`
	NewName        *string `json:"new_name,omitempty" jsonschema:"(Optional) The new name of the file. If unset, the file will be copied with its original name."`
}

type MoveAndRenameItemArgs struct {
	DriveID     string  `json:"drive_id" jsonschema:"(Required) The ID of the drive to update the file in."`
	ItemID      string  `json:"item_id" jsonschema:"(Required) The ID of the item to update."`
	NewFolderID *string `json:"new_folder_id,omitempty" jsonschema:"(Optional) The ID of the folder to move the file to."`
	NewName     *string `json:"new_name,omitempty" jsonschema:"(Optional) The new name of the file."`
}

type AddPermissionArgs struct {
	DriveID            string  `json:"drive_id" jsonschema:"(Required) The ID of the drive to add the permission to."`
	ItemID             string  `json:"item_id" jsonschema:"(Required) The ID of the item to add the permission to."`
	Emails             string  `json:"emails" jsonschema:"(Required) Comma-separated list of email addresses to grant permission to (e.g., 'user1@example.com,user2@example.com')"`
	Role               string  `json:"role" jsonschema:"(Required) The role to grant to the users. Must be one of: 'read' (can view), 'write' (can view and edit)."`
	Message            *string `json:"message,omitempty" jsonschema:"(Optional) The message to send to the users."`
	Password           *string `json:"password,omitempty" jsonschema:"(Optional) The password to use for the permission. if unset, the permission will be added without a password."`
	ExpirationDateTime *string `json:"expiration_date_time,omitempty" jsonschema:"(Optional) The expiration date and time of the permission. if unset, the permission will be added without an expiration date. Format: YYYY-MM-DDTHH:MM:SSZ, like 2018-07-15T14:00:00.000Z"`
}

type DeletePermissionArgs struct {
	DriveID      string `json:"drive_id" jsonschema:"(Required) The ID of the drive to delete the permission from."`
	ItemID       string `json:"item_id" jsonschema:"(Required) The ID of the item to delete the permission from."`
	PermissionID string `json:"permission_id" jsonschema:"(Required) The ID of the permission to delete."`
}

type ListPermissionsArgs struct {
	DriveID string `json:"drive_id" jsonschema:"(Required) The ID of the drive to list the permissions from."`
	ItemID  string `json:"item_id" jsonschema:"(Required) The ID of the item to list the permissions from."`
}

type PermissionInfo struct {
	ID                string  `json:"id"`
	Role              string  `json:"role"`
	GrantedTo         *string `json:"granted_to,omitempty"`
	GrantedToIdentity *string `json:"granted_to_identity,omitempty"`
}

// ListAllDrives lists all available drives for the user
func (o *OneDriveMCPServer) ListAllDrives(ctx context.Context, req *mcp.CallToolRequest, args ListAllDrivesArgs) (*mcp.CallToolResult, any, error) {
	var allDrives []DriveInfo

	// Get personal drive
	drive, err := o.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get personal drive: %w", err)
	}

	if drive != nil && drive.GetId() != nil && drive.GetName() != nil {
		owner := "Personal"
		if drive.GetOwner() != nil && drive.GetOwner().GetUser() != nil && drive.GetOwner().GetUser().GetDisplayName() != nil {
			owner = *drive.GetOwner().GetUser().GetDisplayName()
		}

		driveType := "personal"
		if drive.GetDriveType() != nil {
			driveType = *drive.GetDriveType()
		}

		url := ""
		if drive.GetWebUrl() != nil {
			url = *drive.GetWebUrl()
		}

		allDrives = append(allDrives, DriveInfo{
			ID:    *drive.GetId(),
			Name:  *drive.GetName(),
			Type:  driveType,
			Owner: owner,
			URL:   url,
		})
	}

	// Get organization drives
	drives, err := o.client.Drives().Get(ctx, nil)
	if err == nil && drives.GetValue() != nil {
		for _, orgDrive := range drives.GetValue() {
			if orgDrive.GetId() != nil && orgDrive.GetName() != nil {
				owner := "Organization"
				if orgDrive.GetOwner() != nil && orgDrive.GetOwner().GetUser() != nil && orgDrive.GetOwner().GetUser().GetDisplayName() != nil {
					owner = *orgDrive.GetOwner().GetUser().GetDisplayName()
				}

				driveType := "business"
				if orgDrive.GetDriveType() != nil {
					driveType = *orgDrive.GetDriveType()
				}

				url := ""
				if orgDrive.GetWebUrl() != nil {
					url = *orgDrive.GetWebUrl()
				}

				allDrives = append(allDrives, DriveInfo{
					ID:    *orgDrive.GetId(),
					Name:  *orgDrive.GetName(),
					Type:  driveType,
					Owner: owner,
					URL:   url,
				})
			}
		}
	}

	result, err := json.MarshalIndent(allDrives, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal all drives: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// GetDrive gets details about a specific drive
func (o *OneDriveMCPServer) GetDrive(ctx context.Context, req *mcp.CallToolRequest, args GetDriveArgs) (*mcp.CallToolResult, any, error) {
	var drive models.Driveable
	var err error

	if args.DriveID == nil {
		drive, err = o.client.Me().Drive().Get(ctx, nil)
	} else {
		drive, err = o.client.Drives().ByDriveId(*args.DriveID).Get(ctx, nil)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get drive: %w", err)
	}

	if drive == nil || drive.GetId() == nil || drive.GetName() == nil {
		return nil, nil, fmt.Errorf("invalid drive response")
	}

	owner := "Unknown"
	if drive.GetOwner() != nil && drive.GetOwner().GetUser() != nil && drive.GetOwner().GetUser().GetDisplayName() != nil {
		owner = *drive.GetOwner().GetUser().GetDisplayName()
	}

	driveType := "unknown"
	if drive.GetDriveType() != nil {
		driveType = *drive.GetDriveType()
	}

	url := ""
	if drive.GetWebUrl() != nil {
		url = *drive.GetWebUrl()
	}

	driveInfo := DriveInfo{
		ID:    *drive.GetId(),
		Name:  *drive.GetName(),
		Type:  driveType,
		Owner: owner,
		URL:   url,
	}

	result, _ := json.MarshalIndent(driveInfo, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// ListDriveItems lists items in a drive folder
func (o *OneDriveMCPServer) ListDriveItems(ctx context.Context, req *mcp.CallToolRequest, args ListDriveItemsArgs) (*mcp.CallToolResult, any, error) {
	var items models.DriveItemCollectionResponseable
	var err error

	if args.FolderID == nil || *args.FolderID == "root" {
		// Get root item first to get its ID
		root, err := o.client.Drives().ByDriveId(args.DriveID).Root().Get(ctx, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get drive root: %w", err)
		}
		if root.GetId() == nil {
			return nil, nil, fmt.Errorf("failed to retrieve drive root ID")
		}
		// List root items using the root item ID
		items, err = o.client.Drives().ByDriveId(args.DriveID).Items().ByDriveItemId(*root.GetId()).Children().Get(ctx, nil)
	} else {
		// List items in specific folder
		items, err = o.client.Drives().ByDriveId(args.DriveID).Items().ByDriveItemId(*args.FolderID).Children().Get(ctx, nil)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to list drive items: %w", err)
	}

	var driveItems []DriveItemInfo
	if items.GetValue() != nil {
		for _, item := range items.GetValue() {
			if item.GetId() != nil && item.GetName() != nil {
				itemType := "file"
				if item.GetFolder() != nil {
					itemType = "folder"
				}

				var modified, created string
				if item.GetLastModifiedDateTime() != nil {
					modified = item.GetLastModifiedDateTime().Format(time.RFC3339)
				}
				if item.GetCreatedDateTime() != nil {
					created = item.GetCreatedDateTime().Format(time.RFC3339)
				}

				url := ""
				if item.GetWebUrl() != nil {
					url = *item.GetWebUrl()
				}

				parentPath := ""
				if item.GetParentReference() != nil && item.GetParentReference().GetPath() != nil {
					parentPath = *item.GetParentReference().GetPath()
				}

				driveItem := DriveItemInfo{
					ID:         *item.GetId(),
					Name:       *item.GetName(),
					Type:       itemType,
					Size:       item.GetSize(),
					Modified:   modified,
					Created:    created,
					URL:        url,
					ParentPath: parentPath,
				}

				driveItems = append(driveItems, driveItem)
			}
		}
	}

	result, _ := json.MarshalIndent(driveItems, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// ListSharedWithMeItems lists items shared with the user
func (o *OneDriveMCPServer) ListSharedWithMeItems(ctx context.Context, req *mcp.CallToolRequest, args ListSharedWithMeItemsArgs) (*mcp.CallToolResult, any, error) {
	// Get user's drive first
	drive, err := o.client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user drive: %w", err)
	}

	if drive.GetId() == nil {
		return nil, nil, fmt.Errorf("failed to retrieve user's drive ID")
	}

	items, err := o.client.Drives().ByDriveId(*drive.GetId()).SharedWithMe().GetAsSharedWithMeGetResponse(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list shared items: %w", err)
	}

	var driveItems []DriveItemInfo
	if items.GetValue() != nil {
		for _, item := range items.GetValue() {
			if item.GetId() != nil && item.GetName() != nil {
				itemType := "file"
				if item.GetFolder() != nil {
					itemType = "folder"
				}

				var modified, created string
				if item.GetLastModifiedDateTime() != nil {
					modified = item.GetLastModifiedDateTime().Format(time.RFC3339)
				}
				if item.GetCreatedDateTime() != nil {
					created = item.GetCreatedDateTime().Format(time.RFC3339)
				}

				url := ""
				if item.GetWebUrl() != nil {
					url = *item.GetWebUrl()
				}

				driveItem := DriveItemInfo{
					ID:       *item.GetId(),
					Name:     *item.GetName(),
					Type:     itemType,
					Size:     item.GetSize(),
					Modified: modified,
					Created:  created,
					URL:      url,
				}

				driveItems = append(driveItems, driveItem)
			}
		}
	}

	result, _ := json.MarshalIndent(driveItems, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// GetDriveItem gets details about a specific drive item
func (o *OneDriveMCPServer) GetDriveItem(ctx context.Context, req *mcp.CallToolRequest, args GetDriveItemArgs) (*mcp.CallToolResult, any, error) {
	item, err := o.client.Drives().ByDriveId(args.DriveID).Items().ByDriveItemId(args.ItemID).Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get drive item: %w", err)
	}

	if item == nil || item.GetId() == nil || item.GetName() == nil {
		return nil, nil, fmt.Errorf("invalid drive item response")
	}

	itemType := "file"
	if item.GetFolder() != nil {
		itemType = "folder"
	}

	var modified, created string
	if item.GetLastModifiedDateTime() != nil {
		modified = item.GetLastModifiedDateTime().Format(time.RFC3339)
	}
	if item.GetCreatedDateTime() != nil {
		created = item.GetCreatedDateTime().Format(time.RFC3339)
	}

	url := ""
	if item.GetWebUrl() != nil {
		url = *item.GetWebUrl()
	}

	parentPath := ""
	if item.GetParentReference() != nil && item.GetParentReference().GetPath() != nil {
		parentPath = *item.GetParentReference().GetPath()
	}

	driveItem := DriveItemInfo{
		ID:         *item.GetId(),
		Name:       *item.GetName(),
		Type:       itemType,
		Size:       item.GetSize(),
		Modified:   modified,
		Created:    created,
		URL:        url,
		ParentPath: parentPath,
	}

	result, _ := json.MarshalIndent(driveItem, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
}

// CreateFolder creates a new folder
func (o *OneDriveMCPServer) CreateFolder(ctx context.Context, req *mcp.CallToolRequest, args CreateFolderArgs) (*mcp.CallToolResult, any, error) {
	// Create folder request body
	driveItem := models.NewDriveItem()
	driveItem.SetName(&args.FolderName)
	folder := models.NewFolder()
	driveItem.SetFolder(folder)

	var err error
	if args.FolderID == "root" {
		// Get root item first to get its ID
		root, err := o.client.Drives().ByDriveId(args.DriveID).Root().Get(ctx, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get drive root: %w", err)
		}
		if root.GetId() == nil {
			return nil, nil, fmt.Errorf("failed to retrieve drive root ID")
		}
		_, err = o.client.Drives().ByDriveId(args.DriveID).Items().ByDriveItemId(*root.GetId()).Children().Post(ctx, driveItem, nil)
	} else {
		_, err = o.client.Drives().ByDriveId(args.DriveID).Items().ByDriveItemId(args.FolderID).Children().Post(ctx, driveItem, nil)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Folder \"%s\" created successfully", args.FolderName),
			},
		},
	}, nil, nil
}

// DeleteDriveItem deletes a drive item
func (o *OneDriveMCPServer) DeleteDriveItem(ctx context.Context, req *mcp.CallToolRequest, args DeleteDriveItemArgs) (*mcp.CallToolResult, any, error) {
	err := o.client.Drives().ByDriveId(args.DriveID).Items().ByDriveItemId(args.ItemID).Delete(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete item: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Item deleted successfully",
			},
		},
	}, nil, nil
}

// CopyDriveItem copies a drive item
func (o *OneDriveMCPServer) CopyDriveItem(ctx context.Context, req *mcp.CallToolRequest, args CopyDriveItemArgs) (*mcp.CallToolResult, any, error) {
	// Validate required parameters
	if args.SourceDriveID == "" || args.SourceItemID == "" || args.TargetDriveID == "" || args.TargetFolderID == "" {
		return nil, nil, fmt.Errorf("source and target Drive/Item IDs cannot be empty")
	}

	// Build the parent reference
	parentRef := models.NewItemReference()
	parentRef.SetDriveId(&args.TargetDriveID)
	parentRef.SetId(&args.TargetFolderID)

	// Build the copy request body
	copyBody := drives.NewItemItemsItemCopyPostRequestBody()
	copyBody.SetParentReference(parentRef)
	if args.NewName != nil && *args.NewName != "" {
		copyBody.SetName(args.NewName)
	}

	// Perform the copy operation
	_, err := o.client.
		Drives().
		ByDriveId(args.SourceDriveID).
		Items().
		ByDriveItemId(args.SourceItemID).
		Copy().
		Post(ctx, copyBody, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to copy item %s from drive %s to drive %s: %w", args.SourceItemID, args.SourceDriveID, args.TargetDriveID, err)
	}

	resultText := fmt.Sprintf("Successfully copied item %s from drive %s to drive %s", args.SourceItemID, args.SourceDriveID, args.TargetDriveID)
	if args.NewName != nil && *args.NewName != "" {
		resultText += fmt.Sprintf(" with new name: %s", *args.NewName)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: resultText,
			},
		},
	}, nil, nil
}

// MoveAndRenameItem moves and renames a drive item
func (o *OneDriveMCPServer) MoveAndRenameItem(ctx context.Context, req *mcp.CallToolRequest, args MoveAndRenameItemArgs) (*mcp.CallToolResult, any, error) {
	driveItem := models.NewDriveItem()

	if args.NewName != nil {
		driveItem.SetName(args.NewName)
	}

	if args.NewFolderID != nil {
		parentRef := models.NewItemReference()
		parentRef.SetId(args.NewFolderID)
		driveItem.SetParentReference(parentRef)
	}

	_, err := o.client.Drives().ByDriveId(args.DriveID).Items().ByDriveItemId(args.ItemID).Patch(ctx, driveItem, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update item: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Item updated successfully",
			},
		},
	}, nil, nil
}

// AddPermission adds permission to a drive item
func (o *OneDriveMCPServer) AddPermission(ctx context.Context, req *mcp.CallToolRequest, args AddPermissionArgs) (*mcp.CallToolResult, any, error) {
	if args.DriveID == "me" {
		return nil, nil, fmt.Errorf("drive_id must be the actual drive ID, cannot be 'me'")
	}

	// Validate role
	validRoles := []string{"read", "write"}
	validRole := false
	for _, role := range validRoles {
		if args.Role == role {
			validRole = true
			break
		}
	}
	if !validRole {
		return nil, nil, fmt.Errorf("invalid role: %s, valid roles are: read, write", args.Role)
	}

	// Parse emails
	emails := strings.Split(strings.ReplaceAll(strings.TrimSpace(args.Emails), " ", ""), ",")
	var validEmails []string
	for _, email := range emails {
		if email != "" {
			validEmails = append(validEmails, email)
		}
	}

	if len(validEmails) == 0 {
		return nil, nil, fmt.Errorf("no valid emails provided")
	}

	// Create invitation request
	invite := drives.NewItemItemsItemInvitePostRequestBody()

	// Set recipients
	var recipients []models.DriveRecipientable
	for _, email := range validEmails {
		recipient := models.NewDriveRecipient()
		recipient.SetEmail(&email)
		recipients = append(recipients, recipient)
	}
	invite.SetRecipients(recipients)
	requireSignIn := true
	invite.SetRequireSignIn(&requireSignIn)
	sendInvitation := true
	invite.SetSendInvitation(&sendInvitation)

	// Set roles
	roles := []string{args.Role}
	invite.SetRoles(roles)

	// Set optional parameters
	if args.Message != nil && *args.Message != "" {
		invite.SetMessage(args.Message)
	}

	// Send invitation
	_, err := o.client.Drives().ByDriveId(args.DriveID).Items().ByDriveItemId(args.ItemID).Invite().Post(ctx, invite, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add permissions: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully added permissions for item %s in drive %s\nPermission has been added to the following emails: %s", args.ItemID, args.DriveID, strings.Join(validEmails, ", ")),
			},
		},
	}, nil, nil
}

// DeletePermission deletes a permission from a drive item
func (o *OneDriveMCPServer) DeletePermission(ctx context.Context, req *mcp.CallToolRequest, args DeletePermissionArgs) (*mcp.CallToolResult, any, error) {
	if args.DriveID == "me" {
		return nil, nil, fmt.Errorf("drive_id must be the actual drive ID, cannot be 'me'")
	}

	err := o.client.Drives().ByDriveId(args.DriveID).Items().ByDriveItemId(args.ItemID).Permissions().ByPermissionId(args.PermissionID).Delete(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete permission: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully deleted permission %s from item %s in drive %s", args.PermissionID, args.ItemID, args.DriveID),
			},
		},
	}, nil, nil
}

// ListPermissions lists all permissions for a drive item
func (o *OneDriveMCPServer) ListPermissions(ctx context.Context, req *mcp.CallToolRequest, args ListPermissionsArgs) (*mcp.CallToolResult, any, error) {
	if args.DriveID == "me" {
		return nil, nil, fmt.Errorf("drive_id must be the actual drive ID, cannot be 'me'")
	}

	permissions, err := o.client.Drives().ByDriveId(args.DriveID).Items().ByDriveItemId(args.ItemID).Permissions().Get(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	var permissionInfos []PermissionInfo
	if permissions.GetValue() != nil {
		for _, perm := range permissions.GetValue() {
			if perm.GetId() != nil {
				permInfo := PermissionInfo{
					ID: *perm.GetId(),
				}

				if perm.GetRoles() != nil && len(perm.GetRoles()) > 0 {
					permInfo.Role = perm.GetRoles()[0]
				}

				if perm.GetGrantedTo() != nil && perm.GetGrantedTo().GetUser() != nil {
					if displayName := perm.GetGrantedTo().GetUser().GetDisplayName(); displayName != nil {
						permInfo.GrantedTo = displayName
					}
				}

				permissionInfos = append(permissionInfos, permInfo)
			}
		}
	}

	result, _ := json.MarshalIndent(permissionInfos, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(result),
			},
		},
	}, nil, nil
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
			server := mcp.NewServer(&mcp.Implementation{Name: "onedrive-mcp-server"}, nil)
			return server
		}

		oneDriveServer, err := NewOneDriveMCPServer(token)
		if err != nil {
			log.Printf("Failed to create OneDrive MCP server: %v", err)
			// Return a server that will fail gracefully
			server := mcp.NewServer(&mcp.Implementation{Name: "onedrive-mcp-server"}, nil)
			return server
		}

		server := mcp.NewServer(&mcp.Implementation{Name: "onedrive-mcp-server"}, nil)

		// Create JSON schemas for the tools
		getDriveSchema, _ := jsonschema.For[GetDriveArgs](nil)
		listDriveItemsSchema, _ := jsonschema.For[ListDriveItemsArgs](nil)
		getDriveItemSchema, _ := jsonschema.For[GetDriveItemArgs](nil)
		createFolderSchema, _ := jsonschema.For[CreateFolderArgs](nil)
		deleteDriveItemSchema, _ := jsonschema.For[DeleteDriveItemArgs](nil)
		copyDriveItemSchema, _ := jsonschema.For[CopyDriveItemArgs](nil)
		moveAndRenameItemSchema, _ := jsonschema.For[MoveAndRenameItemArgs](nil)
		addPermissionSchema, _ := jsonschema.For[AddPermissionArgs](nil)
		deletePermissionSchema, _ := jsonschema.For[DeletePermissionArgs](nil)
		listPermissionsSchema, _ := jsonschema.For[ListPermissionsArgs](nil)

		// Register all tools with proper schemas
		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_all_drives",
			Description: "Lists all available OneDrive drives for a user, including the user's personal drive and shared organization drives.",
		}, oneDriveServer.ListAllDrives)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "get_drive",
			Description: "Gets details about a specific OneDrive drive.",
			InputSchema: getDriveSchema,
		}, oneDriveServer.GetDrive)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_drive_items",
			Description: "Lists items under a specific folder or path in a user's OneDrive.",
			InputSchema: listDriveItemsSchema,
		}, oneDriveServer.ListDriveItems)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_shared_with_me_items",
			Description: "Lists all files and folders that have been shared with the user by others.",
		}, oneDriveServer.ListSharedWithMeItems)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "get_drive_item",
			Description: "Gets details about a specific file in a user's OneDrive.",
			InputSchema: getDriveItemSchema,
		}, oneDriveServer.GetDriveItem)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "create_folder",
			Description: "Creates a new folder under a specific folder or path in a user's OneDrive.",
			InputSchema: createFolderSchema,
		}, oneDriveServer.CreateFolder)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "delete_drive_item",
			Description: "Deletes a file from a user's OneDrive.",
			InputSchema: deleteDriveItemSchema,
		}, oneDriveServer.DeleteDriveItem)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "copy_drive_item",
			Description: "Copies a file from one location to another in a user's OneDrive.",
			InputSchema: copyDriveItemSchema,
		}, oneDriveServer.CopyDriveItem)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "move_and_rename_item",
			Description: "Moves and renames a file in a user's OneDrive.",
			InputSchema: moveAndRenameItemSchema,
		}, oneDriveServer.MoveAndRenameItem)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "add_permission",
			Description: "Grants permission to users via email for a specific file in OneDrive.",
			InputSchema: addPermissionSchema,
		}, oneDriveServer.AddPermission)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "delete_permission",
			Description: "Removes a specific permission from a file in OneDrive.",
			InputSchema: deletePermissionSchema,
		}, oneDriveServer.DeletePermission)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "list_permissions",
			Description: "Lists all permissions for a specific file in OneDrive.",
			InputSchema: listPermissionsSchema,
		}, oneDriveServer.ListPermissions)

		return server
	}

	if *httpAddr != "" {
		mcpHandler := mcp.NewStreamableHTTPHandler(serverFactory, nil)
		log.Printf("OneDrive MCP server listening at %s", *httpAddr)

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
