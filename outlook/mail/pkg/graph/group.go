package graph

import (
	"context"
	"fmt"
	// "strings"

	// "github.com/gptscript-ai/tools/outlook/mail/pkg/util"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

// func ListGroupMessages(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, start, end string, limit int) ([]models.ConversationThreadable, error) {
// 	queryParams := &groups.ItemThreadsRequestBuilderGetQueryParameters{
// 		Orderby: []string{"lastDeliveredDateTime DESC"},
// 	}

// 	if limit > 0 {
// 		queryParams.Top = util.Ptr(int32(limit))
// 	}

// 	var filters []string
// 	if start != "" {
// 		filters = append(filters, fmt.Sprintf("lastDeliveredDateTime ge %s", start))
// 	}
// 	if end != "" {
// 		filters = append(filters, fmt.Sprintf("lastDeliveredDateTime le %s", end))
// 	}

// 	if len(filters) > 0 {
// 		queryParams.Filter = util.Ptr(strings.Join(filters, " and "))
// 	}

// 	// Fetch messages from the group mailbox
// 	result, err := client.Groups().ByGroupId(groupID).Threads().Get(ctx, &groups.ItemThreadsRequestBuilderGetRequestConfiguration{
// 		QueryParameters: queryParams,
// 	})

// 	if err != nil {
// 		return nil, fmt.Errorf("failed to list group mailbox messages: %w", err)
// 	}

// 	return result.GetValue(), nil
// }

// ListGroups retrieves all Microsoft 365 groups the authenticated user has access to
func ListGroups(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) ([]models.Groupable, error) {
	// Define query parameters (e.g., sorting or filtering if needed)
	queryParams := &groups.GroupsRequestBuilderGetQueryParameters{
		Select: []string{"id", "displayName", "mail"},
		Orderby: []string{"displayName ASC"},
	}

	// Fetch list of groups
	result, err := client.Groups().Get(ctx, &groups.GroupsRequestBuilderGetRequestConfiguration{
		QueryParameters: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	return result.GetValue(), nil
}