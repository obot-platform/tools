package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/obot-platform/tools/openai-model-provider/api"
)

type upstreamErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func handleValidationError(err error, msg string) error {
	log.Printf("ERROR Invalid: %v", err)
	if msg == "" && err != nil {
		msg = err.Error()
	}
	errorJSON := map[string]string{"error": msg}
	if encErr := json.NewEncoder(os.Stdout).Encode(errorJSON); encErr != nil {
		return encErr
	}
	return errors.New(msg)
}

func parseUpstreamErrorMessage(body []byte) string {
	var upstreamErr upstreamErrorResponse
	if err := json.Unmarshal(body, &upstreamErr); err != nil {
		return ""
	}
	if strings.Contains(upstreamErr.Error.Message, "Incorrect API key provided") || // OpenAI error
		strings.Contains(upstreamErr.Error.Message, "invalid x-api-key") { // Anthropic error
		return "Invalid API Key"
	}
	return upstreamErr.Error.Message
}

func (cfg *Config) Validate() error {
	if err := cfg.EnsureURL(); err != nil {
		return handleValidationError(err, "")
	}

	url := cfg.URL.JoinPath("/models")

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return handleValidationError(err, "")
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "application/json")

	if cfg.RewriteHeaderFn != nil {
		cfg.RewriteHeaderFn(req.Header)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleValidationError(err, "")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := []byte(nil)
		if resp.Body != nil {
			if bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096)); readErr == nil {
				body = bodyBytes
			}
		}

		// Log the status and full body, but only include upstreamMsg or simplified status in the user-facing error
		bodyErr := fmt.Errorf("status %s: %s", resp.Status, string(body))
		msg := parseUpstreamErrorMessage(body)
		if msg == "" {
			msg = fmt.Sprintf("status: %s", resp.Status)
		}
		return handleValidationError(bodyErr, msg)
	}

	var modelsResp api.ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return handleValidationError(err, "Invalid Response Format")
	}

	if modelsResp.Object != "" && modelsResp.Object != "list" || len(modelsResp.Data) == 0 {
		return handleValidationError(nil, fmt.Sprintf("Invalid Models Response: %d models", len(modelsResp.Data)))
	}

	return nil
}
