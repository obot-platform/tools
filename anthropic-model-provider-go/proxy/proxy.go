package proxy

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

const AnthropicBaseHostName = "api.anthropic.com"

type Server struct {
	cfg *proxy.Config
}

func NewServer(cfg *proxy.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) AnthropicProxyRedirect(req *http.Request) {
	req.URL.Scheme = s.cfg.URL.Scheme
	req.URL.Host = s.cfg.URL.Host
	req.URL.Path = s.cfg.URL.JoinPath(strings.TrimPrefix(req.URL.Path, "/v1")).Path
	req.Host = req.URL.Host

	req.Header.Del("Authorization")
	req.Header.Set("x-api-key", s.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	if req.Body == nil || s.cfg.URL.Host != AnthropicBaseHostName || req.URL.Path != proxy.ChatCompletionsPath {
		return
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("failed to read request body, error: ", err.Error())
		return
	}

	var reqBody openai.ChatCompletionRequest
	if err := json.Unmarshal(bodyBytes, &reqBody); err == nil && needsModification(reqBody) {
		if err := modifyRequestBody(req, &reqBody); err != nil {
			fmt.Println("failed to modify request body for claude, error: ", err.Error())
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	} else {
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
}

type ThinkingConfig struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

type ThinkingRequestBody struct {
	openai.ChatCompletionRequest
	Thinking ThinkingConfig `json:"thinking,omitempty"`
}

func isThinkingModel(model string) bool {
	return strings.HasSuffix(model, "-thinking")
}

func needsModification(reqBody openai.ChatCompletionRequest) bool {
	return isThinkingModel(reqBody.Model)
}

func modifyRequestBody(req *http.Request, reqBody *openai.ChatCompletionRequest) error {
	var modifiedBodyBytes []byte
	var err error
	if isThinkingModel(reqBody.Model) {
		reqBody.Model = strings.TrimSuffix(reqBody.Model, "-thinking") // remove our custom -thinking suffix
		reqBody.MaxTokens = 64000                                      // set max tokens to 64000, which is the current max for 3.7 Sonnet in extended thinking mode
		temp := float32(1)
		reqBody.Temperature = &temp
		thinkingReqBody := &ThinkingRequestBody{
			ChatCompletionRequest: *reqBody,
			Thinking: ThinkingConfig{
				Type:         "enabled",
				BudgetTokens: 64000 / 2, // TODO: is 50% of max tokens a good default?
			},
		}
		modifiedBodyBytes, err = json.Marshal(thinkingReqBody)
		if err != nil {
			return err
		}
	} else {
		modifiedBodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body after modification: %w", err)
		}
	}
	req.Body = io.NopCloser(bytes.NewBuffer(modifiedBodyBytes))
	req.ContentLength = int64(len(modifiedBodyBytes))
	return nil
}
