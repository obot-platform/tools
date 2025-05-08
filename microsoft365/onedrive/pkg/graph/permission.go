package graph

import (
	"context"
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func ListDriveItemPermissions(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID string, itemID string) ([]models.Permissionable, error) {
	if driveID == "" || itemID == "" {
		return nil, fmt.Errorf("driveID and itemID cannot be empty")
	}
	permissions, err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(itemID).Permissions().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions for item %s in drive %s: %w", itemID, driveID, err)
	}
	return permissions.GetValue(), nil
}

func DeleteDriveItemPermission(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID string, itemID string, permissionID string) error {
	if driveID == "" || itemID == "" || permissionID == "" {
		return fmt.Errorf("driveID, itemID, and permissionID cannot be empty")
	}
	err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(itemID).Permissions().ByPermissionId(permissionID).Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete permission %s for item %s in drive %s: %w", permissionID, itemID, driveID, err)
	}
	return nil
}

func AddDriveItemPermission(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID string, itemID string, emails []string, role string, message string, password string, expirationDateTime string) ([]models.Permissionable, error) {
	if driveID == "" || itemID == "" || len(emails) == 0 || role == "" {
		return nil, fmt.Errorf("driveID, itemID, emails, and role cannot be empty")
	}
	requestBody := drives.NewItemItemsItemInvitePostRequestBody()
	var recipients []models.DriveRecipientable
	for _, email := range emails {
		driveRecipient := models.NewDriveRecipient()
		driveRecipient.SetEmail(&email)
		recipients = append(recipients, driveRecipient)
	}
	requestBody.SetRecipients(recipients)
	requestBody.SetMessage(&message)
	requireSignIn := true
	requestBody.SetRequireSignIn(&requireSignIn)
	sendInvitation := true
	requestBody.SetSendInvitation(&sendInvitation)
	roles := []string{
		role,
	}
	requestBody.SetRoles(roles)

	if password != "" {
		requestBody.SetPassword(&password)
	}

	if expirationDateTime != "" {
		requestBody.SetExpirationDateTime(&expirationDateTime)
	}

	invite, err := client.Drives().ByDriveId(driveID).Items().ByDriveItemId(itemID).Invite().PostAsInvitePostResponse(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to add permission for item %s in drive %s: %w", itemID, driveID, err)
	}
	return invite.GetValue(), nil
}
