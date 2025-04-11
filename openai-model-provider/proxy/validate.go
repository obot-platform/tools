package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/obot-platform/tools/openai-model-provider/api"
)

func handleValidationError(err error, loggerPath, msg string) error {
	slog.Error(msg, "logger", loggerPath, "error", err)
	fmt.Printf("{\"error\": \"%s\"}\n", msg)
	return fmt.Errorf(msg)
}

func (cfg *Config) Validate(toolPath string) error {
	if err := cfg.EnsureURL(); err != nil {
		return fmt.Errorf("failed to ensure URL: %w", err)
	}

	url := cfg.URL.JoinPath("/models")

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return handleValidationError(err, toolPath, fmt.Sprintf("Invalid %s Configuration", cfg.Name))
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "application/json")

	if cfg.RewriteHeaderFn != nil {
		cfg.RewriteHeaderFn(req.Header)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleValidationError(err, toolPath, fmt.Sprintf("Invalid %s Configuration", cfg.Name))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.Body != nil {
			body, _ := io.ReadAll(resp.Body)
			return handleValidationError(fmt.Errorf("status %s: %s", resp.Status, string(body)), toolPath, fmt.Sprintf("Invalid %s Credentials", cfg.Name))
		}
		return handleValidationError(fmt.Errorf("status %s", resp.Status), toolPath, fmt.Sprintf("Invalid %s Credentials", cfg.Name))
	}

	var modelsResp api.ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return handleValidationError(err, toolPath, "Invalid Response Format")
	}

	if modelsResp.Object != "" && modelsResp.Object != "list" || len(modelsResp.Data) == 0 {
		return handleValidationError(nil, toolPath, fmt.Sprintf("Invalid Models Response: %d models", len(modelsResp.Data)))
	}

	return nil
}
