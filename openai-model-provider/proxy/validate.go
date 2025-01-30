package proxy

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/obot-platform/tools/openai-model-provider/api"
)

func handleValidationError(loggerPath, msg string) error {
	slog.Error(msg, "logger", loggerPath)
	fmt.Printf("{\"error\": \"%s\"}\n", msg)
	return fmt.Errorf(msg)
}

func (cfg *Config) Validate(toolPath string) error {
	if err := cfg.ensureURL(); err != nil {
		return fmt.Errorf("failed to ensure URL: %w", err)
	}

	url := cfg.URL.JoinPath("/models")

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return handleValidationError(toolPath, fmt.Sprintf("Invalid %s Configuration", cfg.Name))
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return handleValidationError(toolPath, fmt.Sprintf("Invalid %s Configuration", cfg.Name))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleValidationError(toolPath, fmt.Sprintf("Invalid %s Credentials", cfg.Name))
	}

	var modelsResp api.ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return handleValidationError(toolPath, "Invalid Response Format")
	}

	if modelsResp.Object != "list" || len(modelsResp.Data) == 0 {
		return handleValidationError(toolPath, "Invalid Models Response")
	}

	return nil
}
