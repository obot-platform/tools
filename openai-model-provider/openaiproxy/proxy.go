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
	return nil
}

func isModelO1(model string) bool {
	if model == "o1" {
		return true
	}
	return strings.HasPrefix(model, "o1-") && !strings.HasPrefix(model, "o1-mini") && !strings.HasPrefix(model, "o1-preview")
}
