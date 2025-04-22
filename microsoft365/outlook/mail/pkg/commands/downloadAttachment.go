package commands

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	"github.com/obot-platform/obot/apiclient"
	"github.com/obot-platform/tools/microsoft365/outlook/common/id"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/client"
	"github.com/obot-platform/tools/microsoft365/outlook/mail/pkg/global"
)

type DownloadAttachmentOpts struct {
	// Message fields
	MessageID string

	// Group message fields
	GroupID       string
	PostID        string
	GroupThreadID string

	// Obot client parameters
	ThreadID    string
	ProjectID   string
	AssistantID string
}

func DownloadAttachment(ctx context.Context, attachmentID string, obotClient *apiclient.Client, opts *DownloadAttachmentOpts) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	var requestInfo *abstractions.RequestInformation
	if opts.GroupID != "" && opts.GroupThreadID != "" && opts.PostID != "" {
		// Group thread post attachment
		requestInfo, err = c.Groups().
			ByGroupId(opts.GroupID).
			Threads().
			ByConversationThreadId(opts.GroupThreadID).
			Posts().
			ByPostId(opts.PostID).
			Attachments().
			ByAttachmentId(attachmentID).
			ToGetRequestInformation(ctx, nil)
	} else {
		// Regular message attachment
		var trueMessageID string
		trueMessageID, err = id.GetOutlookID(ctx, opts.MessageID)
		if err != nil {
			return fmt.Errorf("failed to get outlook ID: %w", err)
		}

		requestInfo, err = c.Me().
			Messages().
			ByMessageId(trueMessageID).
			Attachments().
			ByAttachmentId(attachmentID).
			ToGetRequestInformation(ctx, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to create request info: %w", err)
	}

	response, err := c.GetAdapter().SendPrimitive(ctx, requestInfo, "[]byte", nil)
	if err != nil {
		return fmt.Errorf("failed to get attachment: %w", err)
	}

	rawContent, ok := response.([]byte)
	if !ok {
		return fmt.Errorf("failed to cast response to byte slice")
	}

	var data map[string]interface{}
	err = json.Unmarshal(rawContent, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal attachment content: %w", err)
	}

	contentBytes, ok := data["contentBytes"].(string)
	if !ok {
		return fmt.Errorf("failed to get content bytes from attachment")
	}

	rawContent, err = base64.StdEncoding.DecodeString(contentBytes)
	if err != nil {
		return fmt.Errorf("failed to decode attachment content: %w", err)
	}

	name, ok := data["name"].(string)
	if !ok {
		return fmt.Errorf("failed to get name from attachment")
	}

	if err := obotClient.UploadThreadFile(ctx, opts.ProjectID, opts.ThreadID, opts.AssistantID, name, rawContent); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	if err := obotClient.UploadKnowledgeFile(ctx, opts.ProjectID, opts.ThreadID, opts.AssistantID, name, rawContent); err != nil {
		return fmt.Errorf("failed to upload knowledge file: %w", err)
	}

	fmt.Printf("Successfully downloaded attachment '%s'\n", name)
	return nil
}
