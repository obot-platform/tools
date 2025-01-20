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
	SetCookie         string     `json:"setCookie"`
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

		var setCookie string
		if state.IsExpired() {
			setCookie, err = refreshToken(p, reqObj)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to refresh token: %v", err), http.StatusForbidden)
				return
			}
		}

		ss := SerializableState{
			ExpiresOn:         state.ExpiresOn,
			AccessToken:       state.AccessToken,
			PreferredUsername: state.PreferredUsername,
			User:              state.User,
			Email:             state.Email,
			SetCookie:         setCookie,
		}

		if err = json.NewEncoder(w).Encode(ss); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode state: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func refreshToken(p *oauth2proxy.OAuthProxy, r *http.Request) (string, error) {
	w := &response{
		headers: make(http.Header),
	}

	req, err := http.NewRequest(r.Method, "/oauth2/auth", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh request object: %v", err)
	}

	req.Header = r.Header
	p.ServeHTTP(w, req)

	switch w.status {
	case http.StatusOK, http.StatusAccepted:
		return w.headers.Get("Set-Cookie"), nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return "", fmt.Errorf("refreshing token returned %d: %s", w.status, w.body)
	default:
		return "", fmt.Errorf("refreshing token returned unexpected status %d: %s", w.status, w.body)
	}
}

type response struct {
	headers http.Header
	body    []byte
	status  int
}

func (r *response) Header() http.Header {
	return r.headers
}

func (r *response) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}

func (r *response) WriteHeader(status int) {
	r.status = status
}
