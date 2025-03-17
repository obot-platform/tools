package commands

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader"
	"github.com/gptscript-ai/tools/outlook/common/id"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/pkoukk/tiktoken-go"
)

func GetAttachment(ctx context.Context, messageID, attachmentID string) error {
	trueMessageID, err := id.GetOutlookID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Get attachment as a Parsable object
	requestInfo, err := c.Me().Messages().ByMessageId(trueMessageID).Attachments().ByAttachmentId(attachmentID).ToGetRequestInformation(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create request info: %w", err)
	}

	// Execute request
	response, err := c.BaseRequestBuilder.RequestAdapter.SendPrimitive(ctx, requestInfo, "[]byte", nil)
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

	filetype := data["contentType"].(string)

	loader := documentloader.DefaultDocLoaderFunc(filetype, documentloader.DefaultDocLoaderFuncOpts{})

	docs, err := loader(ctx, bytes.NewReader(rawContent))
	if err != nil {
		return fmt.Errorf("failed to load documents from attachment: %w", err)
	}

	if len(docs) == 0 {
		return fmt.Errorf("no data parsed from attachment")
	}

	var texts []string
	for _, doc := range docs {
		if len(doc.Content) == 0 {
			continue
		}
		texts = append(texts, doc.Content)
	}

	result := strings.Join(texts, "\n---docbreak---\n")

	// Check if text is too large by counting tokens
	tokenizer, err := tiktoken.EncodingForModel("gpt-4")
	if err != nil {
		return fmt.Errorf("failed to create tokenizer: %w", err)
	}

	tokens := tokenizer.Encode(result, nil, nil)
	if len(tokens) > 10000 {
		return fmt.Errorf("attachment content is too large (over 10k tokens)")
	}

	fmt.Println(result)
	return nil
}
