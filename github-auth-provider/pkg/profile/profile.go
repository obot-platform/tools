package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type githubResponse struct {
	AvatarURL string `json:"avatar_url"`
}

func FetchGitHubProfileIconURL(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, result)
	}

	var profile githubResponse
	if err = json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return "", err
	}

	return profile.AvatarURL, nil
}
