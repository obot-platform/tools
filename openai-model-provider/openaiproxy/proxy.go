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

	apiKey := s.cfg.APIKey
	if requestAPIKey := req.Header.Get("X-Obot-OBOT_OPENAI_MODEL_PROVIDER_API_KEY"); requestAPIKey != "" {
		apiKey = requestAPIKey
		req.Header.Del("X-Obot-OBOT_OPENAI_MODEL_PROVIDER_API_KEY")
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	if req.Body == nil || s.cfg.URL.Host != proxy.OpenaiBaseHostName || req.URL.Path != proxy.ChatCompletionsPath {
		return
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("failed to read request body, error: ", err.Error())
		return
	}

	var reqBody openai.ChatCompletionRequest
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return
	}

	if reqBody.Stream {
		if reqBody.StreamOptions == nil {
			reqBody.StreamOptions = new(openai.StreamOptions)
		}
		reqBody.StreamOptions.IncludeUsage = true
	}

	if isModelO1(reqBody.Model) {
		modifyRequestBodyForO1(&reqBody)
	}

	modifiedBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("failed to modify request body: ", err.Error())
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		req.Body = io.NopCloser(bytes.NewBuffer(modifiedBodyBytes))
		req.ContentLength = int64(len(modifiedBodyBytes))
	}
}

func modifyRequestBodyForO1(reqBody *openai.ChatCompletionRequest) {
	reqBody.Temperature = nil
	for i, msg := range reqBody.Messages {
		if msg.Role == "system" {
			reqBody.Messages[i].Role = "developer"
		}
	}
}

func isModelO1(model string) bool {
	if model == "o1" {
		return true
	}
	return strings.HasPrefix(model, "o1-") && !strings.HasPrefix(model, "o1-mini") && !strings.HasPrefix(model, "o1-preview")
}
