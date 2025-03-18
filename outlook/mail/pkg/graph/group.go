package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func PrintThreadMessages(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, threadID string) {
	// Fetch messages inside a thread
	result, err := client.Groups().ByGroupId(groupID).Threads().ByConversationThreadId(threadID).Posts().Get(ctx, nil)
	if err != nil {
		fmt.Printf("❌ Error fetching messages in thread %s: %v\n", threadID, err)
		return
	}

	posts := result.GetValue()
	if len(posts) == 0 {
		fmt.Println("📭 No messages found in this thread.")
		return
	}

	fmt.Println("\n✉️ Messages:")
	for _, post := range posts {
		fmt.Println("------------------------------------------")
		messageID := util.Deref(post.GetId())
		fmt.Printf("📧 Message ID: %s\n", messageID)

		// Check if sender information is available
		if post.GetFrom() != nil && post.GetFrom().GetEmailAddress() != nil {
			fmt.Printf("👤 From: %s <%s>\n",
				util.Deref(post.GetFrom().GetEmailAddress().GetName()),
				util.Deref(post.GetFrom().GetEmailAddress().GetAddress()),
			)
		} else {
			fmt.Println("👤 Sender: Unknown")
		}

		fmt.Printf("📅 Sent: %s\n", post.GetReceivedDateTime().String())

		// Print message body if available
		if post.GetBody() != nil && post.GetBody().GetContent() != nil {
			fmt.Println("📝 Message Body:")
			fmt.Println(util.Deref(post.GetBody().GetContent()))
		} else {
			fmt.Println("📭 (No content in this message)")
		}
		fmt.Println("------------------------------------------")
	}
}

func ListGroupThreads(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, start, end string, limit int) ([]models.ConversationThreadable, error) {
	queryParams := &groups.ItemThreadsRequestBuilderGetQueryParameters{
		Orderby: []string{"lastDeliveredDateTime DESC"},
	}

	if limit > 0 {
		queryParams.Top = util.Ptr(int32(limit))
	}

	var filters []string
	if start != "" {
		filters = append(filters, fmt.Sprintf("lastDeliveredDateTime ge %s", start))
	}
	if end != "" {
		filters = append(filters, fmt.Sprintf("lastDeliveredDateTime le %s", end))
	}

	if len(filters) > 0 {
		queryParams.Filter = util.Ptr(strings.Join(filters, " and "))
	}

	// Fetch messages from the group mailbox
	result, err := client.Groups().ByGroupId(groupID).Threads().Get(ctx, &groups.ItemThreadsRequestBuilderGetRequestConfiguration{
		QueryParameters: queryParams,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list group mailbox messages: %w", err)
	}

	return result.GetValue(), nil
}

// ListGroups retrieves all Microsoft 365 groups the authenticated user has access to
func ListGroups(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) ([]models.Groupable, error) {

	// Fetch groups where the user is a member
	result, err := client.Me().MemberOf().Get(ctx, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to list user groups: %w", err)
	}

	// Filter for groups that have a mailbox (mailEnabled == true)
	var accessibleGroups []models.Groupable
	for _, group := range result.GetValue() {
		if g, ok := group.(models.Groupable); ok {
			if g.GetMailEnabled() != nil && *g.GetMailEnabled() {
				accessibleGroups = append(accessibleGroups, g)
			}
		}
	}

	return accessibleGroups, nil
}

func getGroup(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID string) (models.Groupable, error) {
	groups, err := client.Groups().ByGroupId(groupID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	return groups, nil
}


func CreateGroupThreadMessage(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID string, info DraftInfo) (models.ConversationThreadable, error) {

	requestBody := models.NewConversationThread()
	requestBody.SetTopic(util.Ptr(info.Subject)) 

	post := models.NewPost()
	body := models.NewItemBody()
	body.SetContentType(util.Ptr(models.HTML_BODYTYPE)) 
	body.SetContent(util.Ptr(info.Body)) 
	post.SetBody(body)

	if len(info.Recipients) > 0 {
		post.SetNewParticipants(emailAddressesToRecipientable(info.Recipients))
	} else {
		// if no recipients are provided, use the group email address
		group, err := getGroup(ctx, client, groupID)
		if err != nil {
			return nil, fmt.Errorf("failed to get group: %w", err)
		}
		post.SetNewParticipants(emailAddressesToRecipientable([]string{util.Deref(group.GetMail())}))
	}

	// if len(info.CC) > 0 {
	// 	post.SetCcRecipients(emailAddressesToRecipientable(info.CC))
	// }

	// if len(info.BCC) > 0 {
	// 	post.SetBccRecipients(emailAddressesToRecipientable(info.BCC))
	// }

	posts := []models.Postable {
		post,
	}
	requestBody.SetPosts(posts)

	threads, err := client.Groups().ByGroupId(groupID).Threads().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create group thread message: %w", err)
	}

	// for _, file := range info.Attachments {
	// 	if file == "" {
	// 		return nil, fmt.Errorf("attachment file path cannot be empty")
	// 	}
	// }



	// if len(info.Attachments) > 0 {
	// 	if err := attachFiles(ctx, client, util.Deref(draft.GetId()), info.Attachments); err != nil {
	// 		return nil, fmt.Errorf("failed to attach files to draft: %w", err)
	// 	}
	// }

	return threads, nil
}
