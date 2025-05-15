package graph

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/util"
)

func convertShareLinkToID(shareLink string) string {
	// See https://learn.microsoft.com/en-us/graph/api/shares-get?view=graph-rest-1.0&tabs=http#encoding-sharing-urls
	encodedLink := base64.StdEncoding.EncodeToString([]byte(shareLink))
	encodedLink = strings.TrimRight(encodedLink, "=")
	encodedLink = strings.ReplaceAll(encodedLink, "/", "_")
	encodedLink = strings.ReplaceAll(encodedLink, "+", "-")
	return fmt.Sprintf("u!%s", encodedLink)
}

func GetOneDriveShareLink(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, shareLink string) (string, []byte, error) {
	id := convertShareLinkToID(shareLink)

	item, err := client.Shares().BySharedDriveItemId(id).DriveItem().Get(ctx, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get shared drive item: %w", err)
	}

	content, err := client.Shares().BySharedDriveItemId(id).DriveItem().Content().Get(ctx, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get content of shared drive item: %w", err)
	}

	fileName := util.Deref(item.GetName())
	if fileName == "" {
		return "", nil, fmt.Errorf("shared drive item has no file name")
	}

	return fileName, content, nil
}
