package validate

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const loggerPath = "/tools/deepseek-model-provider/validate"

func init() {
	log.SetFlags(0)
}

func logInfo(msg string) {
	log.Printf("time=%q level=info msg=%q logger=%s", time.Now().Format(time.RFC3339), msg, loggerPath)
}

func logError(msg string, args ...any) {
	log.Printf("time=%q level=error msg=%q %s logger=%s", time.Now().Format(time.RFC3339), msg, fmt.Sprint(args...), loggerPath)
}

func Run(apiKey string) error {
	if err := validateAPIKey(apiKey); err != nil {
		return err
	}
	logInfo("Credentials are valid")
	return nil
}

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

func validateAPIKey(apiKey string) error {
	req, err := http.NewRequest("GET", "https://api.deepseek.com/v1/models", nil)
	if err != nil {
		logError("Failed to create request", fmt.Sprintf("error=%q", err))
		return fmt.Errorf("failed to initialize validation")
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logError("Failed to make request", fmt.Sprintf("error=%q", err))
		return fmt.Errorf("failed to connect to DeepSeek API")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logError("Failed to read response body", fmt.Sprintf("error=%q", err))
		return fmt.Errorf("failed to process API response")
	}

	if resp.StatusCode != 200 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			logError("Authentication failed",
				fmt.Sprintf("error=%q type=%s", errResp.Error.Message, errResp.Error.Type))
			return fmt.Errorf("authentication failed")
		}
		logError("Unexpected status code",
			fmt.Sprintf("status=%d body=%q", resp.StatusCode, string(body)))
		return fmt.Errorf("API validation failed")
	}

	var modelsResp ModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		logError("Failed to parse response",
			fmt.Sprintf("error=%q body=%q", err, string(body)))
		return fmt.Errorf("failed to process API response")
	}

	if len(modelsResp.Data) == 0 {
		logError("No models found in response", fmt.Sprintf("body=%q", string(body)))
		return fmt.Errorf("invalid API response")
	}

	return nil
}

func PrintError(msg string) {
	json.NewEncoder(os.Stdout).Encode(map[string]string{
		"error": msg,
	})
}
