package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	gopenai "github.com/gptscript-ai/chat-completion-client"
	"github.com/obot-platform/tools/openai-model-provider/openaiproxy"
	"github.com/obot-platform/tools/openai-model-provider/proxy"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
	"github.com/openai/openai-go/shared/constant"
)

func main() {
	apiKey := os.Getenv("OBOT_OPENAI_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("OBOT_OPENAI_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	cfg := &proxy.Config{
		APIKey:          apiKey,
		ListenPort:      port,
		BaseURL:         "https://api.openai.com/v1",
		RewriteModelsFn: proxy.DefaultRewriteModelsResponse,
		Name:            "OpenAI",
		CustomPathHandleFuncs: map[string]http.HandlerFunc{
			"/v1/": translateResponsesAPI(apiKey),
		},
	}

	openaiProxy := openaiproxy.NewServer(cfg)
	reverseProxy := &httputil.ReverseProxy{
		Director: openaiProxy.Openaiv1ProxyRedirect,
	}
	cfg.CustomPathHandleFuncs["/v1/"] = reverseProxy.ServeHTTP

	if len(os.Args) > 1 && os.Args[1] == "validate" {
		if err := cfg.Validate("/tools/openai-model-provider/validate"); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := proxy.Run(cfg); err != nil {
		panic(err)
	}
}

type responsesRequestTranslator struct {
	apiKey        string
	wasTranslated bool
}

func translateResponsesAPI(apiKey string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		r := &responsesRequestTranslator{apiKey: apiKey}
		(&httputil.ReverseProxy{
			Director:       r.openaiProxyWithComputerUse,
			ModifyResponse: r.modifyResponsesAPIResponse,
		}).ServeHTTP(rw, req)
	}
}

func (r *responsesRequestTranslator) openaiProxyWithComputerUse(req *http.Request) {
	req.URL.Scheme = "https"
	req.URL.Host = "api.openai.com"
	req.Host = req.URL.Host
	req.Body, req.URL.Path, r.wasTranslated = rewriteBody(req.Body, req.URL.Path)

	req.Header.Set("Authorization", "Bearer "+r.apiKey)
}

func rewriteBody(body io.ReadCloser, path string) (io.ReadCloser, string, bool) {
	if body == nil || path != proxy.ChatCompletionsPath {
		return body, path, false
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return body, path, false
	}

	var chatCompletionRequest gopenai.ChatCompletionRequest
	if err := json.Unmarshal(bodyBytes, &chatCompletionRequest); err != nil || !strings.HasPrefix(chatCompletionRequest.Model, "computer-use-") {
		// Best effort, just return the original body and path on error.
		return io.NopCloser(bytes.NewBuffer(bodyBytes)), path, false
	}

	var (
		text         responses.ResponseTextConfigParam
		inputItems   []responses.ResponseInputItemUnionParam
		tools        []responses.ToolUnionParam
		instructions string
	)
	// Translate the response format
	if chatCompletionRequest.ResponseFormat != nil {
		switch chatCompletionRequest.ResponseFormat.Type {
		case gopenai.ChatCompletionResponseFormatTypeText:
			text = responses.ResponseTextConfigParam{
				Format: responses.ResponseFormatTextConfigUnionParam{
					OfText: &shared.ResponseFormatTextParam{
						Type: constant.Text(gopenai.ChatCompletionResponseFormatTypeText),
					},
				},
			}
		case gopenai.ChatCompletionResponseFormatTypeJSONObject:
			text = responses.ResponseTextConfigParam{
				Format: responses.ResponseFormatTextConfigUnionParam{
					OfJSONObject: &shared.ResponseFormatJSONObjectParam{
						Type: constant.JSONObject(gopenai.ChatCompletionResponseFormatTypeJSONObject),
					},
				},
			}
		default:
			// Best effort log and move on.
			fmt.Fprintln(os.Stderr, "Unsupported response format type:", chatCompletionRequest.ResponseFormat.Type)
		}
	}
	// Translate the initial system message to instructions
	if len(chatCompletionRequest.Messages) > 0 && (chatCompletionRequest.Messages[0].Role == gopenai.ChatMessageRoleSystem || chatCompletionRequest.Messages[0].Role == "developer") {
		instructions = chatCompletionRequest.Messages[0].Content
		chatCompletionRequest.Messages = chatCompletionRequest.Messages[1:]
	}
	// Translate the messages to input items
	inputItems = make([]responses.ResponseInputItemUnionParam, 0, len(chatCompletionRequest.Messages))
	for _, message := range chatCompletionRequest.Messages {
		switch {
		case len(message.ToolCalls) > 0:
			for _, call := range message.ToolCalls {
				inputItems = append(inputItems, responses.ResponseInputItemParamOfFunctionCall(
					call.Function.Arguments,
					call.ID,
					call.Function.Name,
				))
			}
		case message.Role == gopenai.ChatMessageRoleFunction:
			inputItems = append(inputItems, responses.ResponseInputItemParamOfFunctionCallOutput(
				message.ToolCallID,
				message.Content,
			))
		case message.Role == gopenai.ChatMessageRoleUser || message.Role == gopenai.ChatMessageRoleAssistant:
			inputItems = append(inputItems, responses.ResponseInputItemParamOfMessage(
				message.Content,
				responses.EasyInputMessageRole(message.Role),
			))
		default:
			// Best effort log and move on.
			fmt.Fprintln(os.Stderr, "Unsupported message role:", message.Role)
		}
	}
	// Translate the tools to tool union params
	var parameters map[string]any
	for _, tool := range chatCompletionRequest.Tools {
		parameters, _ = tool.Function.Parameters.(map[string]any)
		tools = append(tools, responses.ToolParamOfFunction(
			tool.Function.Name,
			parameters,
			true,
		))
	}
	// Translate the chat completion request to a responses API request
	responsesRequest := responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: inputItems,
		},
		Model: shared.ResponsesModel(chatCompletionRequest.Model),
		Instructions: param.Opt[string]{
			Value: instructions,
		},
		MaxOutputTokens: param.Opt[int64]{
			Value: int64(chatCompletionRequest.MaxTokens),
		},
		ParallelToolCalls: param.Opt[bool]{
			Value: true,
		},
		PreviousResponseID: param.Opt[string]{
			Value: "",
		},
		Store: param.Opt[bool]{
			Value: false,
		},
		Temperature: param.Opt[float64]{
			Value: float64(*chatCompletionRequest.Temperature),
		},
		TopP: param.Opt[float64]{
			Value: float64(chatCompletionRequest.TopP),
		},
		User: param.Opt[string]{
			Value: chatCompletionRequest.User,
		},
		Reasoning:  shared.ReasoningParam{},
		Include:    nil,
		Metadata:   nil,
		Truncation: responses.ResponseNewParamsTruncationDisabled,
		Text:       text,
		ToolChoice: responses.ResponseNewParamsToolChoiceUnion{
			OfToolChoiceMode: param.Opt[responses.ToolChoiceOptions]{
				Value: responses.ToolChoiceOptionsAuto,
			},
		},
		Tools: tools,
	}

	// Marshal the responses request to JSON
	responsesRequestBytes, err := json.Marshal(responsesRequest)
	if err != nil {
		// Best effort, just return the original body and path on error.
		return io.NopCloser(bytes.NewBuffer(bodyBytes)), path, false
	}

	// Return the new body and path
	return io.NopCloser(bytes.NewBuffer(responsesRequestBytes)), "/v1/responses", true
}

func (r *responsesRequestTranslator) modifyResponsesAPIResponse(resp *http.Response) error {
	if r.wasTranslated || resp.StatusCode != http.StatusOK {
		return nil
	}

	return nil
}
