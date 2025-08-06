package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchGitHubProfile(t *testing.T) {
	// Arrange: setup a mock server
	mockResponse := githubUserProfile{
		Login:     "mockuser",
		AvatarURL: "https://example.com/avatar.png",
		Name:      "Mock User",
		Email:     "mockuser@example.com",
		ID:        42,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate Authorization header
		expectedAuth := "Bearer mocktoken"
		if got := r.Header.Get("Authorization"); got != expectedAuth {
			http.Error(w, fmt.Sprintf("unexpected auth header: got %s", got), http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Act: call the function with mock server URL
	githubBaseURL = server.URL
	ctx := context.Background()
	profile, err := FetchUserProfile(ctx, "Bearer mocktoken")

	// Assert: check no error
	if err != nil {
		t.Fatalf("FetchGitHubProfile returned error: %v", err)
	}

	// Assert: check correct values
	if profile.Login != "mockuser" {
		t.Errorf("unexpected login: got %s, want %s", profile.Login, "mockuser")
	}
	if profile.ID != 42 {
		t.Errorf("unexpected ID: got %d, want %d", profile.ID, 42)
	}
	if profile.Name != "Mock User" {
		t.Errorf("unexpected name: got %s, want %s", profile.Name, "Mock User")
	}
	if profile.Email != "mockuser@example.com" {
		t.Errorf("unexpected email: got %s, want %s", profile.Email, "mockuser@example.com")
	}
	if profile.AvatarURL != "https://example.com/avatar.png" {
		t.Errorf("unexpected avatar URL: got %s, want %s", profile.AvatarURL, "https://example.com/avatar.png")
	}
}
