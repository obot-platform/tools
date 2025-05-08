package graph

import (
	"context"
	"fmt"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func ListOrgDrives(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) ([]models.Driveable, error) {
	drives, err := client.Drives().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list drives: %w", err)
	}

	return drives.GetValue(), nil
}

func GetMyDrive(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) (models.Driveable, error) {
	drive, err := client.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get my drive: %w", err)
	}
	return drive, nil
}

func GetDrive(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, driveID string) (models.Driveable, error) {
	if driveID == "" {
		return nil, fmt.Errorf("drive ID cannot be empty")
	}

	drive, err := client.Drives().ByDriveId(driveID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get drive %s: %w", driveID, err)
	}

	return drive, nil
}
