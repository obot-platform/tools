package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchGoogleProfile(t *testing.T) {
	// Arrange: mock response struct
	mockProfile := googleProfile{
		ID:            "123456789",
		Email:         "test@example.com",
		VerifiedEmail: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Picture:       "http://example.com/avatar.jpg",
		HD:            "example.com",
	}

	// Arrange: mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedAuth := "Bearer mock_google_token"
		if got := r.Header.Get("Authorization"); got != expectedAuth {
			http.Error(w, fmt.Sprintf("unexpected auth header: got %s", got), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockProfile)
	}))
	defer server.Close()

	// Act: call your function with mock server URL
	ctx := context.Background()
	profile, err := FetchGoogleProfile(ctx, "Bearer mock_google_token", server.URL)

	// Assert: no error
	if err != nil {
		t.Fatalf("FetchGoogleProfile returned error: %v", err)
	}

	// Assert: check fields
	if profile.ID != mockProfile.ID {
		t.Errorf("unexpected ID: got %s, want %s", profile.ID, mockProfile.ID)
	}
	if profile.Email != mockProfile.Email {
		t.Errorf("unexpected Email: got %s, want %s", profile.Email, mockProfile.Email)
	}
	if profile.VerifiedEmail != mockProfile.VerifiedEmail {
		t.Errorf("unexpected VerifiedEmail: got %v, want %v", profile.VerifiedEmail, mockProfile.VerifiedEmail)
	}
	if profile.Name != mockProfile.Name {
		t.Errorf("unexpected Name: got %s, want %s", profile.Name, mockProfile.Name)
	}
	if profile.GivenName != mockProfile.GivenName {
		t.Errorf("unexpected GivenName: got %s, want %s", profile.GivenName, mockProfile.GivenName)
	}
	if profile.FamilyName != mockProfile.FamilyName {
		t.Errorf("unexpected FamilyName: got %s, want %s", profile.FamilyName, mockProfile.FamilyName)
	}
	if profile.Picture != mockProfile.Picture {
		t.Errorf("unexpected Picture: got %s, want %s", profile.Picture, mockProfile.Picture)
	}
	if profile.HD != mockProfile.HD {
		t.Errorf("unexpected HD: got %s, want %s", profile.HD, mockProfile.HD)
	}
}
