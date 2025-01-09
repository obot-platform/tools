package validate

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Run performs the validation of the DeepSeek API key
func Run(apiKey string) error {
	if err := validateAPIKey(apiKey); err != nil {
		return err
	}
	fmt.Println("Credentials are valid")
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
		log.Printf("Error creating request: %v", err)
		return fmt.Errorf("failed to initialize validation")
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return fmt.Errorf("failed to connect to DeepSeek API")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return fmt.Errorf("failed to process API response")
	}

	if resp.StatusCode != 200 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			log.Printf("DeepSeek API error: %s (Type: %s)", errResp.Error.Message, errResp.Error.Type)
			return fmt.Errorf("authentication failed")
		}
		log.Printf("Unexpected status code: %d, body: %s", resp.StatusCode, string(body))
		return fmt.Errorf("API validation failed")
	}

	var modelsResp ModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		log.Printf("Error parsing response: %v, body: %s", err, string(body))
		return fmt.Errorf("failed to process API response")
	}

	if len(modelsResp.Data) == 0 {
		log.Printf("No models found in response: %s", string(body))
		return fmt.Errorf("invalid API response")
	}

	return nil
}

func printError(msg string) {
	json.NewEncoder(os.Stdout).Encode(map[string]string{
		"error": msg,
	})
}
