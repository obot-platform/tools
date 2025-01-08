package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

type ModelsResponse struct {
	Object string `json:"object"`
	Data   []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

func main() {
	apiKey := os.Getenv("OBOT_DEEPSEEK_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		printError("OBOT_DEEPSEEK_MODEL_PROVIDER_API_KEY environment variable not set")
		return
	}

	if err := validateAPIKey(apiKey); err != nil {
		printError(fmt.Sprintf("Invalid DeepSeek Credentials: %s", err.Error()))
		return
	}

	// Print success message
	fmt.Println("Credentials are valid")
}

func validateAPIKey(apiKey string) error {
	req, err := http.NewRequest("GET", "https://api.deepseek.com/v1/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return fmt.Errorf("%s", errResp.Error.Message)
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var modelsResp ModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(modelsResp.Data) == 0 {
		return fmt.Errorf("no models found in response")
	}

	return nil
}

func printError(msg string) {
	json.NewEncoder(os.Stdout).Encode(map[string]string{
		"error": msg,
	})
}
