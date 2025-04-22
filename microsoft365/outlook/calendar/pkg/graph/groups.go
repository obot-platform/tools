package graph

import (
	"context"

	"github.com/obot-platform/tools/microsoft365/outlook/calendar/pkg/util"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func GetGroupNameFromID(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, id string) (string, error) {
	resp, err := client.Groups().ByGroupId(id).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return util.Deref(resp.GetDisplayName()), nil
}
