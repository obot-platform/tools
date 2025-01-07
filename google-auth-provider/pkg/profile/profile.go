package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type googleProfile struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	HD            string `json:"hd"`
}

func FetchGoogleProfileIconURL(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, result)
	}

	var profile googleProfile
	if err = json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return "", err
	}

	return profile.Picture, nil
}
