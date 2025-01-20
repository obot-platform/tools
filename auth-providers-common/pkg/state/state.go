package state

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	oauth2proxy "github.com/oauth2-proxy/oauth2-proxy/v7"
)

type SerializableRequest struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Header map[string][]string `json:"header"`
}

type SerializableState struct {
	ExpiresOn         *time.Time `json:"expiresOn"`
	AccessToken       string     `json:"accessToken"`
	PreferredUsername string     `json:"preferredUsername"`
	User              string     `json:"user"`
	Email             string     `json:"email"`
}

func ObotGetState(p *oauth2proxy.OAuthProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sr SerializableRequest
		if err := json.NewDecoder(r.Body).Decode(&sr); err != nil {
			http.Error(w, fmt.Sprintf("failed to decode request body: %v", err), http.StatusBadRequest)
			return
		}

		reqObj, err := http.NewRequest(sr.Method, sr.URL, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create request object: %v", err), http.StatusBadRequest)
			return
		}

		reqObj.Header = sr.Header

		state, err := p.LoadCookiedSession(reqObj)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to load cookied session: %v", err), http.StatusBadRequest)
			return
		}

		if state == nil {
			http.Error(w, "state is nil", http.StatusInternalServerError)
			return
		}

		ss := SerializableState{
			ExpiresOn:         state.ExpiresOn,
			AccessToken:       state.AccessToken,
			PreferredUsername: state.PreferredUsername,
			User:              state.User,
			Email:             state.Email,
		}

		if err := json.NewEncoder(w).Encode(ss); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode state: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
