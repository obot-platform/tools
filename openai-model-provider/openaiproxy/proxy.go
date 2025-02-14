package openaiproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	openai "github.com/gptscript-ai/chat-completion-client"
	"github.com/obot-platform/tools/openai-model-provider/proxy"
)

type Server struct {
	cfg *proxy.Config
}

func NewServer(cfg *proxy.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Openaiv1ProxyRedirect(req *http.Request) {
	req.URL.Scheme = s.cfg.URL.Scheme
	req.URL.Host = s.cfg.URL.Host
	req.URL.Path = s.cfg.URL.JoinPath(strings.TrimPrefix(req.URL.Path, "/v1")).Path // join baseURL with request path - /v1 must be part of baseURL if it's needed
	req.Host = req.URL.Host

	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	if req.Body == nil || s.cfg.URL.Host != proxy.OpenaiBaseHostName || req.URL.Path != proxy.ChatCompletionsPath {
		return
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("failed to read request body, error: ", err.Error())
		return
	}

	var reqBody openai.ChatCompletionRequest
	if err := json.Unmarshal(bodyBytes, &reqBody); err == nil && isModelO1(reqBody.Model) {
		if err := modifyRequestBodyForO1(req, &reqBody); err != nil {
			fmt.Println("failed to modify request body for o1, error: ", err.Error())
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	} else {
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
}

func modifyRequestBodyForO1(req *http.Request, reqBody *openai.ChatCompletionRequest) error {
	reqBody.Stream = false
	reqBody.Temperature = nil
	for i, msg := range reqBody.Messages {
		if msg.Role == "system" {
			reqBody.Messages[i].Role = "developer"
		}
	}
	modifiedBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body after modification: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(modifiedBodyBytes))
	req.ContentLength = int64(len(modifiedBodyBytes))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "")
	req.Header.Set("Content-Type", "application/json")
	return nil
}

func (s *Server) ModifyResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK || resp.Request.URL.Path != proxy.ChatCompletionsPath || resp.Request.URL.Host != proxy.OpenaiBaseHostName {
		return nil
	}

	if resp.Header.Get("Content-Type") == "application/json" {
		rawBody, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to read response body: %w", err)
		}
		resp.Body.Close()
		var respBody openai.ChatCompletionResponse
		if err := json.Unmarshal(rawBody, &respBody); err == nil && isModelO1(respBody.Model) {
			// Convert non-streaming response to a single SSE for o1 model
			streamResponse := openai.ChatCompletionStreamResponse{
				ID:      respBody.ID,
				Object:  respBody.Object,
				Created: respBody.Created,
				Model:   respBody.Model,
				Usage:   respBody.Usage,
				Choices: func() []openai.ChatCompletionStreamChoice {
					var choices []openai.ChatCompletionStreamChoice
					for _, choice := range respBody.Choices {
						choices = append(choices, openai.ChatCompletionStreamChoice{
							Index: choice.Index,
							Delta: openai.ChatCompletionStreamChoiceDelta{
								Content:      choice.Message.Content,
								Role:         choice.Message.Role,
								FunctionCall: choice.Message.FunctionCall,
								ToolCalls:    choice.Message.ToolCalls,
							},
							FinishReason: choice.FinishReason,
						})
					}
					return choices
				}(),
			}

			sseData, err := json.Marshal(streamResponse)
			if err != nil {
				return fmt.Errorf("failed to marshal stream response: %w", err)
			}

			sseFormattedData := fmt.Sprintf("data: %s\n\nevent: close\ndata: [DONE]\n\n", sseData)

			resp.Header.Set("Content-Type", "text/event-stream")
			resp.Header.Set("Cache-Control", "no-cache")
			resp.Header.Set("Connection", "keep-alive")
			resp.Body = io.NopCloser(bytes.NewBufferString(sseFormattedData))
		} else {
			resp.Body = io.NopCloser(bytes.NewBuffer(rawBody))
		}
	}

	return nil
}

func isModelO1(model string) bool {
	if model == "o1" {
		return true
	}
	return strings.HasPrefix(model, "o1-") && !strings.HasPrefix(model, "o1-mini") && !strings.HasPrefix(model, "o1-preview")
}
