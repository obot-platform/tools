package graph

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/obot-platform/tools/knowledge/pkg/datastore/documentloader"
	"github.com/obot-platform/tools/knowledge/pkg/flows"
	flowconfig "github.com/obot-platform/tools/knowledge/pkg/flows/config"
	"github.com/pkoukk/tiktoken-go"
)

func GetAttachmentContent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, requestInfo *abstractions.RequestInformation) (string, error) {
	response, err := client.GetAdapter().SendPrimitive(ctx, requestInfo, "[]byte", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get attachment: %w", err)
	}

	rawContent, ok := response.([]byte)
	if !ok {
		return "", fmt.Errorf("failed to cast response to byte slice")
	}

	var data map[string]any
	err = json.Unmarshal(rawContent, &data)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal attachment content: %w", err)
	}

	contentBytes, ok := data["contentBytes"].(string)
	if !ok {
		return "", fmt.Errorf("failed to get content bytes from attachment")
	}

	rawContent, err = base64.StdEncoding.DecodeString(contentBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decode attachment content: %w", err)
	}

	filetype := data["contentType"].(string)

	flowCfg, err := flowconfig.Load("blueprint:obot")
	if err != nil {
		return "", fmt.Errorf("failed to load flow config: %w", err)
	}

	var flow *flowconfig.FlowConfigEntry
	flow, err = flowCfg.GetFlow("obotload")
	if err != nil {
		return "", fmt.Errorf("failed to get flow: %w", err)
	}

	var loader documentloader.LoaderFunc
	var ingestionFlow *flows.IngestionFlow
	for _, ingestionFlowConfig := range flow.Ingestion {
		flow, err := ingestionFlowConfig.AsIngestionFlow(&flow.Globals.Ingestion)
		if err != nil {
			return "", fmt.Errorf("failed to create ingestion flow: %w", err)
		}
		if flow.SupportsFiletype(filetype) {
			ingestionFlow = flow
			break
		}
	}

	if ingestionFlow != nil {
		if err := ingestionFlow.FillDefaults(filetype); err != nil {
			return "", fmt.Errorf("failed to fill defaults: %w", err)
		}
		loader = ingestionFlow.Load
	} else {
		loader = documentloader.DefaultDocLoaderFunc(filetype, documentloader.DefaultDocLoaderFuncOpts{})
	}

	docs, err := loader(ctx, bytes.NewReader(rawContent))
	if err != nil {
		return "", fmt.Errorf("failed to load documents from attachment: %w", err)
	}

	if len(docs) == 0 {
		return "", fmt.Errorf("no data parsed from attachment")
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
		return "", fmt.Errorf("failed to create tokenizer: %w", err)
	}

	tokens := tokenizer.Encode(result, nil, nil)
	if len(tokens) > 10000 {
		return "", fmt.Errorf("attachment content is too large (over 10k tokens)")
	}

	return result, nil
}
