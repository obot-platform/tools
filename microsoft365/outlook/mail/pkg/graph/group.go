package graph

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/gptscript-ai/go-gptscript"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	users "github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/util"
)

func ListThreadMessages(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, threadID string) ([]models.Postable, error) {
	// Fetch messages inside a thread
	result, err := client.Groups().ByGroupId(groupID).Threads().ByConversationThreadId(threadID).Posts().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list group mailbox messages: %w", err)
	}

	return result.GetValue(), nil
}

func GetThreadMessage(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, threadID, postID string) (models.Postable, error) {
	// Fetch messages inside a thread
	result, err := client.Groups().ByGroupId(groupID).Threads().ByConversationThreadId(threadID).Posts().ByPostId(postID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get group mailbox message: %w", err)
	}

	return result, nil
}

func ListGroupThreads(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID string, limit int) ([]models.ConversationThreadable, error) {
	queryParams := &groups.ItemThreadsRequestBuilderGetQueryParameters{
		Orderby: []string{"lastDeliveredDateTime DESC"},
	}

	if limit > 0 {
		queryParams.Top = util.Ptr(int32(limit))
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

func GetUserType(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) (string, error) {
	opts := &users.UserItemRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UserItemRequestBuilderGetQueryParameters{
			Select: []string{"displayName", "userType", "userPrincipalName"},
		},
	}
	me, err := client.Me().Get(ctx, opts) // Note: In the future, if we ever updated the SDK to a version that the /me doesn't have userType, we can use the /users/{upn} endpoint to get userType instead
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %v", err)
	}

	userType := me.GetUserType()
	if userType == nil { // Personal accounts don't have userType
		return "Personal", nil
	}

	return *userType, nil
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
	if slices.Contains(info.Attachments, "") {
		return nil, fmt.Errorf("attachment file path cannot be empty")
	}

	requestBody := models.NewConversationThread()
	requestBody.SetTopic(util.Ptr(info.Subject))

	post := models.NewPost()
	body := models.NewItemBody()
	body.SetContentType(util.Ptr(models.HTML_BODYTYPE))
	body.SetContent(util.Ptr(info.Body))
	post.SetBody(body)

	if len(info.Recipients) > 0 {
		post.SetNewParticipants(emailAddressesToRecipientable(info.Recipients))
	}

	// models.Post() doesn't support cc and bcc

	if len(info.Attachments) > 0 {
		attachments, err := setAttachments(ctx, info.Attachments)
		if err != nil {
			return nil, fmt.Errorf("failed to attach files to group thread message post: %w", err)
		}
		post.SetAttachments(attachments)
	}
	posts := []models.Postable{
		post,
	}

	requestBody.SetPosts(posts)

	threads, err := client.Groups().ByGroupId(groupID).Threads().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create group thread message: %w", err)
	}

	return threads, nil
}

func ReplyToGroupThreadMessage(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, threadID string, info DraftInfo) error {
	for _, file := range info.Attachments {
		if file == "" {
			return fmt.Errorf("attachment file path cannot be empty")
		}
	}

	requestBody := groups.NewItemConversationsItemThreadsItemReplyPostRequestBody()
	post := models.NewPost()
	body := models.NewItemBody()
	body.SetContentType(util.Ptr(models.HTML_BODYTYPE))
	body.SetContent(util.Ptr(info.Body))
	post.SetBody(body)

	if len(info.Recipients) > 0 {
		post.SetNewParticipants(emailAddressesToRecipientable(info.Recipients))
	}

	// models.Post() doesn't support cc and bcc

	if len(info.Attachments) > 0 {
		attachments, err := setAttachments(ctx, info.Attachments)
		if err != nil {
			return fmt.Errorf("failed to attach files to group thread message post: %w", err)
		}
		post.SetAttachments(attachments)
	}
	requestBody.SetPost(post)

	err := client.Groups().ByGroupId(groupID).Threads().ByConversationThreadId(threadID).Reply().Post(ctx, requestBody, nil)
	if err != nil {
		return fmt.Errorf("failed to reply to group thread message %s: %w", threadID, err)
	}

	return nil
}

func setAttachments(ctx context.Context, attachment_filenames []string) ([]models.Attachmentable, error) {
	attachments := []models.Attachmentable{}
	gsClient, err := gptscript.NewGPTScript()
	if err != nil {
		return nil, fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	for _, filename := range attachment_filenames {
		attachment := models.NewFileAttachment()
		attachment.SetName(util.Ptr(filename))

		data, err := gsClient.ReadFileInWorkspace(ctx, filepath.Join("files", filename))
		if err != nil {
			return nil, fmt.Errorf("failed to read attachment file %s from workspace: %v", filename, err)
		}

		attachment.SetContentBytes(data)
		attachments = append(attachments, attachment)
	}

	return attachments, nil

}

func DeleteGroupThread(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, threadID string) error {
	err := client.Groups().ByGroupId(groupID).Threads().ByConversationThreadId(threadID).Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete group thread: %w", err)
	}
	return nil
}

func ListGroupThreadMessageAttachments(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, threadID, postID string) ([]models.Attachmentable, error) {
	result, err := client.Groups().
		ByGroupId(groupID).
		Threads().
		ByConversationThreadId(threadID).
		Posts().
		ByPostId(postID).
		Attachments().
		Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list attachments for group thread message: %w", err)
	}

	return result.GetValue(), nil
}

func GetGroupThreadMessageAttachment(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, threadID, postID, attachmentID string) (string, error) {
	// Get attachment as a Parsable object
	requestInfo, err := client.Groups().
		ByGroupId(groupID).
		Threads().
		ByConversationThreadId(threadID).
		Posts().
		ByPostId(postID).
		Attachments().
		ByAttachmentId(attachmentID).
		ToGetRequestInformation(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request info: %w", err)
	}

	result, err := GetAttachmentContent(ctx, client, requestInfo)
	if err != nil {
		return "", fmt.Errorf("failed to get attachment content: %w", err)
	}

	return result, nil
}
